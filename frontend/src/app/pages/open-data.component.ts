import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { TrafficService } from '../core/api.services';
import { ViolationStat, ViolationType } from '../models';

// Javna stranica (open data) — dostupna bez prijave.
// Prikazuje isključivo anonimne, agregirane podatke.
@Component({
  selector: 'app-open-data',
  standalone: true,
  imports: [CommonModule],
  template: `
    <h2>Otvoreni podaci — saobraćajni prekršaji</h2>
    <p class="hint">
      Anonimna statistika saobraćajnih prekršaja.
    </p>

    <div class="card">
      <h3>Statistika prekršaja po tipu</h3>
      @if (stats.length === 0) {
        <p class="hint">Još nema evidentiranih prekršaja.</p>
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

  constructor(private traffic: TrafficService) {}

  ngOnInit(): void {
    this.traffic.getViolationStats().subscribe((s) => (this.stats = s));
    this.traffic.getViolationTypes().subscribe((t) => (this.types = t));
  }

  typeLabel(code: string): string {
    return this.types.find((t) => t.code === code)?.label ?? code;
  }
}
