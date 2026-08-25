import { api } from './client';
import { PlateReservation, Vehicle, VehicleDetails, VehicleReportDetails, VerifyReportResponse } from '../types';

export interface RegisterVehicleInput {
  ownerCitizenId: number;
  vin: string;
  make: string;
  model: string;
  year: number;
  category: string;
  color: string;
  enginePowerKw: number;
  fuelType: string;
  insuranceValidUntil: string;
  techInspectionValidUntil: string;
}

export const vehiclesApi = {
  // ---- sopstveni podaci (građanin) ----
  myVehicles: () => api.get<Vehicle[]>('/me/vehicles'),
  myReservations: () => api.get<PlateReservation[]>('/me/plate-reservations'),

  // ---- registar vozila (službenik) ----
  list: () => api.get<Vehicle[]>('/vehicles'),
  get: (id: number) => api.get<VehicleDetails>(`/vehicles/${id}`),
  register: (input: RegisterVehicleInput) => api.post<Vehicle>('/vehicles', input),
  transfer: (id: number, newOwnerCitizenId: number) =>
    api.post<Vehicle>(`/vehicles/${id}/transfer`, { newOwnerCitizenId }),
  renew: (id: number) => api.post<Vehicle>(`/vehicles/${id}/renew-registration`),

  // ---- krađa/pronalazak (vlasnik ili službenik) ----
  reportTheft: (id: number) => api.post(`/vehicles/${id}/report-theft`),
  reportFound: (id: number) => api.post<Vehicle>(`/vehicles/${id}/report-found`),

  // ---- digitalni izveštaj ----
  generateReport: (id: number) => api.post<VehicleReportDetails>(`/vehicles/${id}/reports`),
  verifyReport: (code: string) => api.get<VerifyReportResponse>(`/reports/verify/${encodeURIComponent(code)}`),

  // ---- personalizovane tablice ----
  requestPlate: (requestedPlate: string) => api.post<PlateReservation>('/plate-reservations', { requestedPlate }),
  listReservations: () => api.get<PlateReservation[]>('/plate-reservations'),
  decideReservation: (id: number, approve: boolean) =>
    api.put<PlateReservation>(`/plate-reservations/${id}/decision`, { approve }),
};
