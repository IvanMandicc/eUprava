import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { NotificationService, PaymentService, TrafficService } from '../../core/api.services';
import { Driver, FineDetails, Violation, ViolationType } from '../../models';

@Component({
  selector: 'app-citizen-dashboard',
  standalone: true,
  imports: [CommonModule],
  template: `
    <h2>Moj vozački karton</h2>

    @if (notDriver) {
      <div class="card">
        <p>Niste evidentirani kao vozač u sistemu saobraćajne policije.</p>
      </div>
    } @else if (driver) {
      <div class="cards-row">
        <div class="card stat">
          <span class="stat-label">Broj vozačke dozvole</span>
          <span class="stat-value">{{ driver.licenseNumber }}</span>
        </div>
        <div class="card stat">
          <span class="stat-label">Kazneni poeni</span>
          <span class="stat-value" [class.danger]="driver.penaltyPoints >= 18">
            {{ driver.penaltyPoints }} / 18
          </span>
        </div>
        <div class="card stat">
          <span class="stat-label">Status dozvole</span>
          <span class="badge" [class.badge-ok]="driver.licenseStatus === 'valid'"
                [class.badge-danger]="driver.licenseStatus === 'suspended'">
            {{ driver.licenseStatus === 'valid' ? 'Važeća' : 'Suspendovana' }}
          </span>
        </div>
      </div>
    }

    <div class="card">
      <h3>Moji prekršaji</h3>
      @if (violations.length === 0) {
        <p class="hint">Nemate evidentiranih prekršaja.</p>
      } @else {
        <table>
          <thead>
            <tr><th>Datum</th><th>Tip</th><th>Mesto</th><th>Poeni</th><th>Kazna (RSD)</th><th>Status</th></tr>
          </thead>
          <tbody>
            @for (v of violations; track v.id) {
              <tr>
                <td>{{ v.date | date: 'dd.MM.yyyy.' }}</td>
                <td>{{ typeLabel(v.type) }}</td>
                <td>{{ v.location }}</td>
                <td>{{ v.points }}</td>
                <td>{{ v.fineAmount | number: '1.2-2' }}</td>
                <td>
                  <span class="badge" [class.badge-warn]="v.status === 'active'" [class.badge-ok]="v.status === 'resolved'">
                    {{ v.status === 'active' ? 'Aktivan' : 'Rešen' }}
                  </span>
                </td>
              </tr>
            }
          </tbody>
        </table>
      }
    </div>

    <div class="card">
      <h3>Moje kazne</h3>
      @if (error) {
        <p class="error">{{ error }}</p>
      }
      @if (fines.length === 0) {
        <p class="hint">Nemate evidentiranih kazni.</p>
      } @else {
        <table>
          <thead>
            <tr><th>Br.</th><th>Prekršaj</th><th>Iznos (RSD)</th><th>Rok za plaćanje</th><th>Status</th><th></th></tr>
          </thead>
          <tbody>
            @for (f of fines; track f.id) {
              <tr>
                <td>{{ f.id }}</td>
                <td>{{ typeLabel(f.violationType) }}</td>
                <td>{{ f.amount | number: '1.2-2' }}</td>
                <td>{{ f.paymentDeadline | date: 'dd.MM.yyyy.' }}</td>
                <td>
                  <span class="badge" [class.badge-ok]="f.paid" [class.badge-warn]="!f.paid">
                    {{ f.paid ? 'Plaćena' : 'Neplaćena' }}
                  </span>
                </td>
                <td>
                  @if (!f.paid) {
                    <button (click)="pay(f)" [disabled]="paying === f.id">
                      {{ paying === f.id ? 'Plaćanje...' : 'Plati' }}
                    </button>
                  }
                </td>
              </tr>
            }
          </tbody>
        </table>
      }
    </div>
  `,
})
export class CitizenDashboardComponent implements OnInit {
  driver: Driver | null = null;
  notDriver = false;
  violations: Violation[] = [];
  fines: FineDetails[] = [];
  types: ViolationType[] = [];
  paying: number | null = null;
  error = '';

  constructor(private traffic: TrafficService, private payments: PaymentService, private notifications: NotificationService) {}

  ngOnInit(): void {
    this.load();
    this.traffic.getViolationTypes().subscribe((t) => (this.types = t));
  }

  load(): void {
    this.traffic.getMyDriver().subscribe({
      next: (d) => (this.driver = d),
      error: () => (this.notDriver = true),
    });
    this.traffic.getMyViolations().subscribe((v) => (this.violations = v));
    this.traffic.getMyFines().subscribe((f) => (this.fines = f));
  }

  typeLabel(code: string): string {
    return this.types.find((t) => t.code === code)?.label ?? code;
  }

  // Plaćanje ide kroz Payment servis: kreiranje pa potvrda (simulacija uplate).
  pay(fine: FineDetails): void {
    this.paying = fine.id;
    this.error = '';
    this.payments.create(fine.id).subscribe({
      next: (p) =>
        this.payments.confirm(p.id).subscribe({
          next: () => {
            this.paying = null;
            this.load();
          },
          error: (err) => this.payError(err),
        }),
      error: (err) => this.payError(err),
    });
  }

  private payError(err: any): void {
    this.error = err.error?.error ?? 'Plaćanje nije uspelo.';
    this.paying = null;
  }
}
