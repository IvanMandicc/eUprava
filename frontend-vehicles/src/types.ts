export type Role = 'citizen' | 'officer' | 'admin';

export interface User {
  id: number;
  jmbg: string;
  firstName: string;
  lastName: string;
  email: string;
  role: Role;
  address: string;
  createdAt: string;
}

export type VehicleStatus = 'registered' | 'stolen';

export interface Vehicle {
  id: number;
  ownerCitizenId: number;
  vin: string;
  plateNumber: string;
  make: string;
  model: string;
  year: number;
  category: string;
  color: string;
  enginePowerKw: number;
  fuelType: string;
  firstRegistrationDate: string;
  insuranceValidUntil: string;
  techInspectionValidUntil: string;
  registrationValidUntil: string;
  status: VehicleStatus;
  createdAt: string;
}

export interface CitizenInfo {
  id: number;
  jmbg: string;
  firstName: string;
  lastName: string;
  email: string;
  role: string;
  address: string;
}

export interface VehicleDetails extends Vehicle {
  owner?: CitizenInfo;
}

export type PlateReservationStatus = 'pending' | 'approved' | 'rejected' | 'expired';

export interface PlateReservation {
  id: number;
  requestedByCitizenId: number;
  requestedPlate: string;
  feeAmount: number;
  status: PlateReservationStatus;
  requestedAt: string;
  decidedAt: string | null;
}

export interface OwnershipTransfer {
  id: number;
  vehicleId: number;
  fromCitizenId: number;
  toCitizenId: number;
  transferDate: string;
}

export interface VehicleReportDetails {
  id: number;
  vehicleId: number;
  verificationCode: string;
  generatedAt: string;
  vehicle: Vehicle;
  transfers: OwnershipTransfer[];
}

export interface VerifyReportResponse {
  valid: boolean;
  report?: VehicleReportDetails;
  error?: string;
}
