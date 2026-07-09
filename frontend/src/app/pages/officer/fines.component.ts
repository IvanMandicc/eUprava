import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { TrafficService } from '../../core/api.services';
import { FineDetails, ViolationType } from '../../models';

@Component({
  selector: 'app-fines',
  standalone: true,
  imports: [CommonModule],
  template: `
    <h2>Sve kazne</h2>
    <div class="card">
      <table>
        <thead>
          <tr><th>Br.</th><th>Prekršaj</th><th>Vozač</th><th>Iznos (RSD)</th><th>Rok</th><th>Status</th><th>Datum plaćanja</th></tr>
        </thead>
        <tbody>
          @for (f of fines; track f.id) {
            <tr>
              <td>{{ f.id }}</td>
              <td>{{ typeLabel(f.violationType) }}</td>
              <td>#{{ f.driverId }}</td>
              <td>{{ f.amount | number: '1.2-2' }}</td>
              <td>{{ f.paymentDeadline | date: 'dd.MM.yyyy.' }}</td>
              <td>
                <span class="badge" [class.badge-ok]="f.paid" [class.badge-warn]="!f.paid">
                  {{ f.paid ? 'Plaćena' : 'Neplaćena' }}
                </span>
              </td>
              <td>{{ f.paymentDate ? (f.paymentDate | date: 'dd.MM.yyyy. HH:mm') : '—' }}</td>
            </tr>
          }
        </tbody>
      </table>
    </div>
  `,
})
export class FinesComponent implements OnInit {
  fines: FineDetails[] = [];
  types: ViolationType[] = [];

  constructor(private traffic: TrafficService) {}

  ngOnInit(): void {
    this.traffic.getFines().subscribe((f) => (this.fines = f));
    this.traffic.getViolationTypes().subscribe((t) => (this.types = t));
  }

  typeLabel(code: string): string {
    return this.types.find((t) => t.code === code)?.label ?? code;
  }
}
