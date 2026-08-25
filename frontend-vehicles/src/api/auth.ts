import { api } from './client';
import { User } from '../types';

export interface RegisterInput {
  jmbg: string;
  firstName: string;
  lastName: string;
  email: string;
  password: string;
  address: string;
}

interface LoginResponse {
  token: string;
  user: User;
}

// Isti /api/auth endpoint kao Angular aplikacija — deljen sistem korisnika
// (SSO): oba frontenda pozivaju istu prijavu i dobijaju token istog formata.
export const authApi = {
  login: (email: string, password: string) => api.post<LoginResponse>('/auth/login', { email, password }),
  register: (input: RegisterInput) => api.post<User>('/auth/register', input),
};
