import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../core/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule, RouterLink],
  template: `
    <div class="card narrow">
      <h2>Prijava</h2>
      <form (ngSubmit)="submit()">
        <label>Email</label>
        <input type="email" name="email" [(ngModel)]="email" required />
        <label>Lozinka</label>
        <input type="password" name="password" [(ngModel)]="password" required />
        @if (error) {
          <p class="error">{{ error }}</p>
        }
        <button type="submit" [disabled]="loading">Prijavi se</button>
      </form>
      <p class="hint">Nemate nalog? <a routerLink="/register">Registrujte se</a></p>
    </div>
  `,
})
export class LoginComponent {
  email = '';
  password = '';
  error = '';
  loading = false;

  constructor(private auth: AuthService, private router: Router) {}

  submit(): void {
    this.loading = true;
    this.error = '';
    this.auth.login(this.email, this.password).subscribe({
      next: (res) =>
        this.router.navigate([
          res.user.role === 'admin' ? '/admin/users' : res.user.role === 'officer' ? '/officer/drivers' : '/citizen',
        ]),
      error: (err) => {
        this.error = err.error?.error ?? 'Prijava nije uspela.';
        this.loading = false;
      },
    });
  }
}
