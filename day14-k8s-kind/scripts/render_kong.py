"""สร้าง kong.yml สำหรับ cluster จากเทมเพลตของ Day 13 + public key ทุกตัวใน keys/*.pub

ทำแบบเดียวกับ day13-revocation-ui-cicd/scripts/render-gateway.ps1 (แทนช่วงระหว่าง marker) แต่เป็น Python
เพื่อให้รันได้ทั้งบนเครื่อง Windows และใน CI โดยไม่ต้องมี PowerShell  Kong ไม่เคยเห็น private key
"""
import pathlib
import re
import sys

BEGIN = "      # __JWT_KEYS_BEGIN__"
END = "# __JWT_KEYS_END__"


def render(template: str, keys: dict) -> str:
    if not keys:
        raise ValueError("no public keys")
    items = []
    for kid in sorted(keys):
        if not re.fullmatch(r"[A-Za-z0-9_-]{1,32}", kid):
            raise ValueError(f"bad kid: {kid}")
        pem = keys[kid].strip().replace("\r", "")
        if not pem.startswith("-----BEGIN PUBLIC KEY-----"):
            raise ValueError(f"{kid}.pub is not a PKIX public key")
        body = "\n".join("          " + line for line in pem.split("\n"))
        items.append(f"      - key: {kid}\n        algorithm: RS256\n        rsa_public_key: |\n{body}")
    pat = re.compile(re.escape(BEGIN) + r".*?" + re.escape(END), re.S)
    if not pat.search(template):
        raise ValueError("template has no __JWT_KEYS_BEGIN__/__JWT_KEYS_END__ markers")
    return pat.sub(lambda _m: "\n".join(items), template)


def main(argv):
    if len(argv) != 4:
        print("usage: render_kong.py <kong.yml.tpl> <keys-dir> <out>", file=sys.stderr)
        return 2
    tpl, keys_dir, out = (pathlib.Path(a) for a in argv[1:])
    keys = {p.stem: p.read_text(encoding="utf-8") for p in sorted(keys_dir.glob("*.pub"))}
    text = render(tpl.read_text(encoding="utf-8"), keys)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(text, encoding="utf-8", newline="\n")
    print("rendered", out, "with kids:", ", ".join(sorted(keys)))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
