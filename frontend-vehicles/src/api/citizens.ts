import { api } from './client';
import { User } from '../types';

export const citizensApi = {
  searchByJmbg: (jmbg: string) => api.get<User>(`/citizens/search?jmbg=${encodeURIComponent(jmbg)}`),
  suggestByJmbg: (prefix: string) => api.get<User[]>(`/citizens/suggest?jmbg=${encodeURIComponent(prefix)}`),
};
