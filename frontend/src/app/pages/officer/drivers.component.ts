import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { CitizenService, TrafficService } from '../../core/api.services';
import { Driver, DriverDetails, User, Violation, ViolationType } from '../../models';

@Component({
  selector: 'app-drivers',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <h2>Evidencija vozača</h2>

    <div class="card">
      <h3>Novi vozač</h3>
      <p class="hint">Pronađi građanina po JMBG-u, pa ga evidentiraj kao vozača.</p>

      <form class="inline-form" (ngSubmit)="searchCitizen()">
        <input name="jmbg" [(ngModel)]="searchJmbg" placeholder="JMBG građanina (13 cifara)" required minlength="13" maxlength="13" />
        <button type="submit">Pretraži</button>
      </form>
      @if (searchError) {
        <p class="error">{{ searchError }}</p>
      }

      @if (foundCitizen) {
        <p class="hint">
          Pronađen: <strong>{{ foundCitizen.firstName }} {{ foundCitizen.lastName }}</strong>
          ({{ foundCitizen.email }}, uloga: {{ foundCitizen.role }})
        </p>
        <form class="inline-form" (ngSubmit)="createDriver()">
          <input name="licenseNumber" [(ngModel)]="newLicenseNumber" placeholder="Broj vozačke dozvole" required />
          <button type="submit">Evidentiraj kao vozača</button>
        </form>
        @if (createError) {
          <p class="error">{{ createError }}</p>
        }
      }
    </div>

    <div class="card">
      <h3>Vozači</h3>
      <table>
        <thead>
          <tr><th>ID</th><th>ID građanina</th><th>Dozvola</th><th>Poeni</th><th>Status</th><th></th></tr>
        </thead>
        <tbody>
          @for (d of drivers; track d.id) {
            <tr>
              <td>{{ d.id }}</td>
              <td>{{ d.citizenId }}</td>
              <td>{{ d.licenseNumber }}</td>
              <td>{{ d.penaltyPoints }}</td>
              <td>
                <span class="badge" [class.badge-ok]="d.licenseStatus === 'valid'" [class.badge-danger]="d.licenseStatus === 'suspended'">
                  {{ d.licenseStatus === 'valid' ? 'Važeća' : 'Suspendovana' }}
                </span>
              </td>
              <td><button class="secondary" (click)="select(d)">Detalji</button></td>
            </tr>
          }
        </tbody>
      </table>
    </div>

    @if (selected) {
      <div class="card">
        <h3>
          Vozač #{{ selected.id }}
          @if (selected.citizen) {
            — {{ selected.citizen.firstName }} {{ selected.citizen.lastName }} (JMBG: {{ selected.citizen.jmbg }})
          }
        </h3>

        <h4>Novi prekršaj</h4>
        <form class="inline-form" (ngSubmit)="createViolation()">
          <select name="type" [(ngModel)]="newViolation.type" required>
            <option value="" disabled>Tip prekršaja</option>
            @for (t of types; track t.code) {
              <option [value]="t.code">{{ t.label }} ({{ t.points }} poena, {{ t.fine }} RSD)</option>
            }
          </select>
          <input name="location" [(ngModel)]="newViolation.location" placeholder="Mesto" required />
          <input name="description" [(ngModel)]="newViolation.description" placeholder="Opis" />
          <button type="submit">Evidentiraj prekršaj</button>
        </form>
        @if (violationError) {
          <p class="error">{{ violationError }}</p>
        }

        <h4>Prekršaji vozača</h4>
        @if (violations.length === 0) {
          <p class="hint">Vozač nema prekršaja.</p>
        } @else {
          <table>
            <thead>
              <tr><th>Datum</th><th>Tip</th><th>Mesto</th><th>Poeni</th><th>Kazna</th><th>Status</th></tr>
            </thead>
            <tbody>
              @for (v of violations; track v.id) {
                <tr>
                  <td>{{ v.date | date: 'dd.MM.yyyy.' }}</td>
                  <td>{{ typeLabel(v.type) }}</td>
                  <td>{{ v.location }}</td>
                  <td>{{ v.points }}</td>
                  <td>{{ v.fineAmount | number: '1.2-2' }}</td>
                  <td>{{ v.status === 'active' ? 'Aktivan' : 'Rešen' }}</td>
                </tr>
              }
            </tbody>
          </table>
        }
      </div>
    }
  `,
})
export class DriversComponent implements OnInit {
  drivers: Driver[] = [];
  selected: DriverDetails | null = null;
  violations: Violation[] = [];
  types: ViolationType[] = [];

  searchJmbg = '';
  foundCitizen: User | null = null;
  searchError = '';
  newLicenseNumber = '';
  createError = '';

  newViolation = { type: '', location: '', description: '' };
  violationError = '';

  constructor(private traffic: TrafficService, private citizens: CitizenService) {}

  ngOnInit(): void {
    this.loadDrivers();
    this.traffic.getViolationTypes().subscribe((t) => (this.types = t));
  }

  loadDrivers(): void {
    this.traffic.getDrivers().subscribe((d) => (this.drivers = d));
  }

  searchCitizen(): void {
    this.searchError = '';
    this.foundCitizen = null;
    this.citizens.searchByJmbg(this.searchJmbg).subscribe({
      next: (u) => (this.foundCitizen = u),
      error: (err) => (this.searchError = err.error?.error ?? 'Građanin nije pronađen.'),
    });
  }

  createDriver(): void {
    if (!this.foundCitizen || !this.newLicenseNumber) return;
    this.createError = '';
    this.traffic.createDriver(this.foundCitizen.id, this.newLicenseNumber).subscribe({
      next: () => {
        this.foundCitizen = null;
        this.searchJmbg = '';
        this.newLicenseNumber = '';
        this.loadDrivers();
      },
      error: (err) => (this.createError = err.error?.error ?? 'Evidentiranje vozača nije uspelo.'),
    });
  }

  select(d: Driver): void {
    this.traffic.getDriver(d.id).subscribe((details) => (this.selected = details));
    this.traffic.getDriverViolations(d.id).subscribe((v) => (this.violations = v));
  }

  createViolation(): void {
    if (!this.selected || !this.newViolation.type) return;
    this.violationError = '';
    this.traffic
      .createViolation({ driverId: this.selected.id, ...this.newViolation })
      .subscribe({
        next: () => {
          this.newViolation = { type: '', location: '', description: '' };
          this.select(this.selected!);
          this.loadDrivers();
        },
        error: (err) => (this.violationError = err.error?.error ?? 'Unos prekršaja nije uspeo.'),
      });
  }

  typeLabel(code: string): string {
    return this.types.find((t) => t.code === code)?.label ?? code;
  }
}
