import { Component } from '@angular/core';
import { ClaimList } from './claim-list/claim-list';

@Component({
  imports: [ClaimList],
  selector: 'app-root',
  styleUrl: './app.css',
  templateUrl: './app.html',
})
export class App {}
