import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { AdminService } from '../../core/api.services';
import { User } from '../../models';

@Component({
  selector: 'app-users',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <h2>Upravljanje korisnicima</h2>

    <div class="card">
      <h3>Novi policajac</h3>
      <p class="hint">
        Javna registracija uvek daje ulogu građanina — nalog policajca može da kreira samo administrator.
      </p>
      <form class="inline-form" (ngSubmit)="createOfficer()">
        <input name="jmbg" [(ngModel)]="officer.jmbg" placeholder="JMBG" required minlength="13" maxlength="13" />
        <input name="firstName" [(ngModel)]="officer.firstName" placeholder="Ime" required />
        <input name="lastName" [(ngModel)]="officer.lastName" placeholder="Prezime" required />
        <input type="email" name="email" [(ngModel)]="officer.email" placeholder="Email" required />
        <input type="password" name="password" [(ngModel)]="officer.password" placeholder="Lozinka" required minlength="6" />
        <button type="submit">Kreiraj policajca</button>
      </form>
      @if (error) {
        <p class="error">{{ error }}</p>
      }
      @if (success) {
        <p class="hint">{{ success }}</p>
      }
    </div>

    <div class="card">
      <h3>Svi korisnici sistema</h3>
      <table>
        <thead>
          <tr><th>ID</th><th>JMBG</th><th>Ime i prezime</th><th>Email</th><th>Uloga</th><th>Registrovan</th></tr>
        </thead>
        <tbody>
          @for (u of users; track u.id) {
            <tr>
              <td>{{ u.id }}</td>
              <td>{{ u.jmbg }}</td>
              <td>{{ u.firstName }} {{ u.lastName }}</td>
              <td>{{ u.email }}</td>
              <td>
                <span class="badge"
                      [class.badge-ok]="u.role === 'citizen'"
                      [class.badge-warn]="u.role === 'officer'"
                      [class.badge-danger]="u.role === 'admin'">
                  {{ roleLabel(u.role) }}
                </span>
              </td>
              <td>{{ u.createdAt | date: 'dd.MM.yyyy.' }}</td>
            </tr>
          }
        </tbody>
      </table>
    </div>
  `,
})
export class UsersComponent implements OnInit {
  users: User[] = [];
  officer = { jmbg: '', firstName: '', lastName: '', email: '', password: '', address: 'MUP Republike Srbije' };
  error = '';
  success = '';

  constructor(private admin: AdminService) {}

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.admin.listUsers().subscribe((u) => (this.users = u));
  }

  createOfficer(): void {
    this.error = '';
    this.success = '';
    this.admin.createOfficer(this.officer).subscribe({
      next: (u) => {
        this.success = `Policajac ${u.firstName} ${u.lastName} je kreiran.`;
        this.officer = { jmbg: '', firstName: '', lastName: '', email: '', password: '', address: 'MUP Republike Srbije' };
        this.load();
      },
      error: (err) => (this.error = err.error?.error ?? 'Kreiranje policajca nije uspelo.'),
    });
  }

  roleLabel(role: string): string {
    return role === 'admin' ? 'Administrator' : role === 'officer' ? 'Policajac' : 'Građanin';
  }
}
