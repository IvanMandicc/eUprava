import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TrafficService } from '../../core/api.services';
import { Violation, ViolationType } from '../../models';

@Component({
  selector: 'app-violations',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <h2>Svi prekršaji</h2>
    <div class="card">
      @if (error) {
        <p class="error">{{ error }}</p>
      }
      <table>
        <thead>
          <tr><th>ID</th><th>Vozač</th><th>Datum</th><th>Tip</th><th>Mesto</th><th>Opis</th><th>Poeni</th><th>Kazna</th><th>Status</th><th></th></tr>
        </thead>
        <tbody>
          @for (v of violations; track v.id) {
            <tr>
              <td>{{ v.id }}</td>
              <td>#{{ v.driverId }}</td>
              <td>{{ v.date | date: 'dd.MM.yyyy.' }}</td>
              <td>{{ typeLabel(v.type) }}</td>
              @if (editingId === v.id) {
                <td><input [(ngModel)]="editLocation" placeholder="Mesto" /></td>
                <td><input [(ngModel)]="editDescription" placeholder="Opis" /></td>
              } @else {
                <td>{{ v.location }}</td>
                <td>{{ v.description }}</td>
              }
              <td>{{ v.points }}</td>
              <td>{{ v.fineAmount | number: '1.2-2' }}</td>
              <td>
                <select [ngModel]="v.status" (ngModelChange)="changeStatus(v, $event)">
                  <option value="active">Aktivan</option>
                  <option value="resolved">Rešen</option>
                </select>
              </td>
              <td class="actions">
                @if (editingId === v.id) {
                  <button (click)="saveEdit(v)">Sačuvaj</button>
                  <button class="secondary" (click)="cancelEdit()">Otkaži</button>
                } @else {
                  <button class="secondary" (click)="startEdit(v)">Izmeni</button>
                  <button class="danger" (click)="remove(v)">Obriši</button>
                }
              </td>
            </tr>
          }
        </tbody>
      </table>
    </div>
  `,
})
export class ViolationsComponent implements OnInit {
  violations: Violation[] = [];
  types: ViolationType[] = [];
  error = '';

  editingId: number | null = null;
  editLocation = '';
  editDescription = '';

  constructor(private traffic: TrafficService) {}

  ngOnInit(): void {
    this.load();
    this.traffic.getViolationTypes().subscribe((t) => (this.types = t));
  }

  load(): void {
    this.traffic.getViolations().subscribe((v) => (this.violations = v));
  }

  typeLabel(code: string): string {
    return this.types.find((t) => t.code === code)?.label ?? code;
  }

  startEdit(v: Violation): void {
    this.editingId = v.id;
    this.editLocation = v.location;
    this.editDescription = v.description;
  }

  cancelEdit(): void {
    this.editingId = null;
  }

  saveEdit(v: Violation): void {
    this.error = '';
    this.traffic
      .updateViolation(v.id, { location: this.editLocation, description: this.editDescription })
      .subscribe({
        next: (updated) => {
          v.location = updated.location;
          v.description = updated.description;
          this.editingId = null;
        },
        error: (err) => (this.error = err.error?.error ?? 'Izmena nije uspela.'),
      });
  }

  changeStatus(v: Violation, status: string): void {
    this.traffic.updateViolation(v.id, { status }).subscribe({
      next: (updated) => (v.status = updated.status),
      error: (err) => (this.error = err.error?.error ?? 'Izmena nije uspela.'),
    });
  }

  remove(v: Violation): void {
    if (!confirm(`Obrisati prekršaj #${v.id}? Vozaču se vraćaju kazneni poeni.`)) return;
    this.error = '';
    this.traffic.deleteViolation(v.id).subscribe({
      next: () => this.load(),
      error: (err) => (this.error = err.error?.error ?? 'Brisanje nije uspelo.'),
    });
  }
}
