import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TrafficService } from '../core/api.services';
import { ViolationStat, ViolationType } from '../models';

// Javna stranica (open data) — dostupna bez prijave.
// Prikazuje isključivo anonimne, agregirane podatke.
@Component({
  selector: 'app-open-data',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <h2>Otvoreni podaci — saobraćajni prekršaji</h2>
    <p class="hint">
      Anonimna statistika saobraćajnih prekršaja.
    </p>

    <div class="card">
      <h3>Statistika prekršaja po tipu</h3>

      <form class="inline-form" (ngSubmit)="applyFilter()">
        <span class="hint">Period (datum prekršaja):</span>
        <input type="date" name="from" [(ngModel)]="from" />
        <span class="hint">—</span>
        <input type="date" name="to" [(ngModel)]="to" />
        <button type="submit">Filtriraj</button>
        <button type="button" class="secondary" (click)="resetFilter()">Resetuj</button>
        <button type="button" class="secondary" (click)="downloadCsv()" [disabled]="stats.length === 0">
          Preuzmi CSV
        </button>
      </form>
      @if (error) {
        <p class="error">{{ error }}</p>
      }

      @if (stats.length === 0) {
        <p class="hint">Nema evidentiranih prekršaja za izabrani period.</p>
      } @else {
        <table>
          <thead>
            <tr><th>Prekršaj</th><th>Broj prekršaja</th><th>Ukupno poena</th><th>Prosečna kazna (RSD)</th><th>Ukupno kazni (RSD)</th></tr>
          </thead>
          <tbody>
            @for (s of stats; track s.type) {
              <tr>
                <td>{{ typeLabel(s.type) }}</td>
                <td>{{ s.count }}</td>
                <td>{{ s.totalPoints }}</td>
                <td>{{ s.avgFine | number: '1.2-2' }}</td>
                <td>{{ s.totalFines | number: '1.2-2' }}</td>
              </tr>
            }
          </tbody>
        </table>
      }
    </div>

    <div class="card">
      <h3>Šifarnik prekršaja (kazneni poeni i iznosi)</h3>
      <table>
        <thead>
          <tr><th>Šifra</th><th>Prekršaj</th><th>Kazneni poeni</th><th>Iznos kazne (RSD)</th></tr>
        </thead>
        <tbody>
          @for (t of types; track t.code) {
            <tr>
              <td>{{ t.code }}</td>
              <td>{{ t.label }}</td>
              <td>{{ t.points }}</td>
              <td>{{ t.fine | number: '1.2-2' }}</td>
            </tr>
          }
        </tbody>
      </table>
    </div>
  `,
})
export class OpenDataComponent implements OnInit {
  stats: ViolationStat[] = [];
  types: ViolationType[] = [];
  from = '';
  to = '';
  error = '';

  constructor(private traffic: TrafficService) {}

  ngOnInit(): void {
    this.loadStats();
    this.traffic.getViolationTypes().subscribe((t) => (this.types = t));
  }

  applyFilter(): void {
    this.error = '';
    if (this.from && this.to && this.from > this.to) {
      this.error = 'Datum "od" ne može biti posle datuma "do".';
      return;
    }
    this.loadStats();
  }

  resetFilter(): void {
    this.from = '';
    this.to = '';
    this.error = '';
    this.loadStats();
  }

  private loadStats(): void {
    this.traffic.getViolationStats(this.from, this.to).subscribe({
      next: (s) => (this.stats = s),
      error: (err) => (this.error = err.error?.error ?? 'Učitavanje statistike nije uspelo.'),
    });
  }

  typeLabel(code: string): string {
    return this.types.find((t) => t.code === code)?.label ?? code;
  }

  // Izvoz trenutno prikazane (filtrirane) statistike kao CSV fajl.
  downloadCsv(): void {
    const header = ['Sifra', 'Prekrsaj', 'Broj prekrsaja', 'Ukupno poena', 'Prosecna kazna (RSD)', 'Ukupno kazni (RSD)'];
    const rows = this.stats.map((s) => [
      s.type,
      this.typeLabel(s.type),
      String(s.count),
      String(s.totalPoints),
      s.avgFine.toFixed(2),
      s.totalFines.toFixed(2),
    ]);
    const csv = [header, ...rows]
      .map((row) => row.map((cell) => `"${cell.replace(/"/g, '""')}"`).join(','))
      .join('\r\n');

    // BOM na početku da Excel ispravno prikaže dijakritike (č, ć, š...).
    const blob = new Blob(['﻿' + csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const period = this.from || this.to ? `_${this.from || 'pocetak'}_${this.to || 'kraj'}` : '';
    const a = document.createElement('a');
    a.href = url;
    a.download = `statistika-prekrsaja${period}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }
}
