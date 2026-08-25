import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { environment } from '../../environments/environment';
import {
  AppNotification,
  Driver,
  DriverDetails,
  Fine,
  FineDetails,
  Payment,
  User,
  Violation,
  ViolationStat,
  ViolationType,
} from '../models';

const API = environment.apiUrl;

// Po jedan Angular servis za svaki mikroservis (preko API Gateway-a).

@Injectable({ providedIn: 'root' })
export class TrafficService {
  constructor(private http: HttpClient) {}

  // ---- rute građanina (svoji podaci) ----
  getMyDriver() {
    return this.http.get<Driver>(`${API}/me/driver`);
  }
  getMyViolations() {
    return this.http.get<Violation[]>(`${API}/me/violations`);
  }
  getMyFines() {
    return this.http.get<FineDetails[]>(`${API}/me/fines`);
  }

  // ---- rute policajca ----
  getDrivers() {
    return this.http.get<Driver[]>(`${API}/drivers`);
  }
  createDriver(citizenId: number, licenseNumber: string) {
    return this.http.post<Driver>(`${API}/drivers`, { citizenId, licenseNumber });
  }
  getDriver(id: number) {
    return this.http.get<DriverDetails>(`${API}/drivers/${id}`);
  }
  getDriverViolations(id: number) {
    return this.http.get<Violation[]>(`${API}/drivers/${id}/violations`);
  }
  getViolations() {
    return this.http.get<Violation[]>(`${API}/violations`);
  }
  createViolation(input: { driverId: number; type: string; description: string; location: string }) {
    return this.http.post<Violation>(`${API}/violations`, input);
  }
  updateViolation(id: number, input: { description?: string; location?: string; status?: string }) {
    return this.http.put<Violation>(`${API}/violations/${id}`, input);
  }
  deleteViolation(id: number) {
    return this.http.delete<{ message: string }>(`${API}/violations/${id}`);
  }
  getFines() {
    return this.http.get<FineDetails[]>(`${API}/fines`);
  }
  getViolationTypes() {
    return this.http.get<ViolationType[]>(`${API}/violation-types`);
  }

  // Open data — javno dostupno bez prijave. from/to su opcioni datumi (GGGG-MM-DD)
  // koji filtriraju statistiku po datumu prekršaja.
  getViolationStats(from?: string, to?: string) {
    let params = new HttpParams();
    if (from) params = params.set('from', from);
    if (to) params = params.set('to', to);
    return this.http.get<ViolationStat[]>(`${API}/open-data/violation-stats`, { params });
  }
}

@Injectable({ providedIn: 'root' })
export class CitizenService {
  constructor(private http: HttpClient) {}

  // Pretraga po JMBG-u — koristi je policajac pri evidentiranju vozača,
  // umesto ručnog unosa internog ID-ja iz baze.
  searchByJmbg(jmbg: string) {
    const params = new HttpParams().set('jmbg', jmbg);
    return this.http.get<User>(`${API}/citizens/search`, { params });
  }
}

@Injectable({ providedIn: 'root' })
export class AdminService {
  constructor(private http: HttpClient) {}

  listUsers() {
    return this.http.get<User[]>(`${API}/users`);
  }
  createOfficer(input: {
    jmbg: string;
    firstName: string;
    lastName: string;
    email: string;
    password: string;
    address: string;
  }) {
    return this.http.post<User>(`${API}/users/officers`, input);
  }
}

@Injectable({ providedIn: 'root' })
export class PaymentService {
  constructor(private http: HttpClient) {}

  create(fineId: number) {
    return this.http.post<Payment>(`${API}/payments`, { fineId });
  }
  confirm(id: number) {
    return this.http.post<Payment>(`${API}/payments/${id}/confirm`, {});
  }
  list() {
    return this.http.get<Payment[]>(`${API}/payments`);
  }
}

@Injectable({ providedIn: 'root' })
export class NotificationService {
  constructor(private http: HttpClient) {}

  list() {
    return this.http.get<AppNotification[]>(`${API}/notifications`);
  }
  markRead(id: number) {
    return this.http.put<{ message: string }>(`${API}/notifications/${id}/read`, {});
  }
}
