locals {
  tags = {
    project = "self-study-2026"
    day     = "07"
    managed = "terraform"
  }
  # public: 10.20.0.0/24, 10.20.1.0/24   private: 10.20.10.0/24, 10.20.11.0/24
  public_cidrs  = [for i, az in var.azs : cidrsubnet(var.vpc_cidr, 8, i)]
  private_cidrs = [for i, az in var.azs : cidrsubnet(var.vpc_cidr, 8, i + 10)]
}

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags                 = merge(local.tags, { Name = "${var.name_prefix}-vpc" })
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags   = merge(local.tags, { Name = "${var.name_prefix}-igw" })
}

resource "aws_subnet" "public" {
  count                   = length(var.azs)
  vpc_id                  = aws_vpc.main.id
  cidr_block              = local.public_cidrs[count.index]
  availability_zone       = var.azs[count.index]
  map_public_ip_on_launch = true
  tags                    = merge(local.tags, { Name = "${var.name_prefix}-public-${count.index}", tier = "public" })
}

resource "aws_subnet" "private" {
  count             = length(var.azs)
  vpc_id            = aws_vpc.main.id
  cidr_block        = local.private_cidrs[count.index]
  availability_zone = var.azs[count.index]
  tags              = merge(local.tags, { Name = "${var.name_prefix}-private-${count.index}", tier = "private" })
}

# public: ออกอินเทอร์เน็ตผ่าน Internet Gateway
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
  tags = merge(local.tags, { Name = "${var.name_prefix}-public-rt" })
}

resource "aws_route_table_association" "public" {
  count          = length(var.azs)
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# private: ไม่มีเส้นทางออกอินเทอร์เน็ต (ไม่สร้าง NAT Gateway เพราะคิดเงินตามชั่วโมง)
# ถ้า service ใน private ต้องเรียกออก ให้ใช้ VPC endpoint หรือเพิ่ม NAT ตอนใช้จริง
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id
  tags   = merge(local.tags, { Name = "${var.name_prefix}-private-rt" })
}

resource "aws_route_table_association" "private" {
  count          = length(var.azs)
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}

# ---------------------------------------------------------------- security groups (chain: internet -> alb -> app -> db)
resource "aws_security_group" "alb" {
  name        = "${var.name_prefix}-alb"
  description = "load balancer: รับ HTTP/HTTPS จากอินเทอร์เน็ต"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = local.tags
}

resource "aws_security_group" "app" {
  name        = "${var.name_prefix}-app"
  description = "api: รับ 8080 จาก ALB เท่านั้น"
  vpc_id      = aws_vpc.main.id

  ingress {
    description     = "from ALB"
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = local.tags
}

resource "aws_security_group" "db" {
  name        = "${var.name_prefix}-db"
  description = "postgres: รับ 5432 จาก app เท่านั้น"
  vpc_id      = aws_vpc.main.id

  ingress {
    description     = "from app"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
  }
  tags = local.tags
}
