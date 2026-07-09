import { HttpClient } from '@angular/common/http';
import { Injectable, computed, signal } from '@angular/core';
import { Router } from '@angular/router';
import { tap } from 'rxjs';
import { environment } from '../../environments/environment';
import { User } from '../models';

interface LoginResponse {
  token: string;
  user: User;
}

export interface RegisterInput {
  jmbg: string;
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  address: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly userSignal = signal<User | null>(this.loadUser());

  readonly user = computed(() => this.userSignal());
  readonly isLoggedIn = computed(() => this.userSignal() !== null);
  readonly isOfficer = computed(() => this.userSignal()?.role === 'officer');
  readonly isCitizen = computed(() => this.userSignal()?.role === 'citizen');
  readonly isAdmin = computed(() => this.userSignal()?.role === 'admin');

  constructor(private http: HttpClient, private router: Router) {}

  login(email: string, password: string) {
    return this.http
      .post<LoginResponse>(`${environment.apiUrl}/auth/login`, { email, password })
      .pipe(
        tap((res) => {
          localStorage.setItem('token', res.token);
          localStorage.setItem('user', JSON.stringify(res.user));
          this.userSignal.set(res.user);
        })
      );
  }

  register(input: RegisterInput) {
    return this.http.post<User>(`${environment.apiUrl}/auth/register`, input);
  }

  logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    this.userSignal.set(null);
    this.router.navigate(['/login']);
  }

  get token(): string | null {
    return localStorage.getItem('token');
  }

  private loadUser(): User | null {
    const raw = localStorage.getItem('user');
    return raw ? (JSON.parse(raw) as User) : null;
  }
}
