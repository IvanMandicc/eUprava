import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService, RegisterInput } from '../core/auth.service';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [FormsModule, RouterLink],
  template: `
    <div class="card narrow">
      <h2>Registracija građanina</h2>
      <form (ngSubmit)="submit()">
        <label>JMBG</label>
        <input name="jmbg" [(ngModel)]="input.jmbg" required minlength="13" maxlength="13" />
        <label>Ime</label>
        <input name="firstName" [(ngModel)]="input.firstName" required />
        <label>Prezime</label>
        <input name="lastName" [(ngModel)]="input.lastName" required />
        <label>Email</label>
        <input type="email" name="email" [(ngModel)]="input.email" required />
        <label>Lozinka (min 6 karaktera)</label>
        <input type="password" name="password" [(ngModel)]="input.password" required minlength="6" />
        <label>Adresa</label>
        <input name="address" [(ngModel)]="input.address" />
        @if (error) {
          <p class="error">{{ error }}</p>
        }
        <button type="submit" [disabled]="loading">Registruj se</button>
      </form>
      <p class="hint">Već imate nalog? <a routerLink="/login">Prijavite se</a></p>
    </div>
  `,
})
export class RegisterComponent {
  input: RegisterInput = { jmbg: '', firstName: '', lastName: '', email: '', password: '', address: '' };
  error = '';
  loading = false;

  constructor(private auth: AuthService, private router: Router) {}

  submit(): void {
    this.loading = true;
    this.error = '';
    this.auth.register(this.input).subscribe({
      next: () => this.router.navigate(['/login']),
      error: (err) => {
        this.error = err.error?.error ?? 'Registracija nije uspela.';
        this.loading = false;
      },
    });
  }
}
