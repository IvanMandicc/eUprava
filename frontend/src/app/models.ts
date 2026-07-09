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

export interface Driver {
  id: number;
  citizenId: number;
  licenseNumber: string;
  penaltyPoints: number;
  licenseStatus: 'valid' | 'suspended';
}

export interface CitizenInfo {
  id: number;
  jmbg: string;
  firstName: string;
  lastName: string;
  email: string;
  address: string;
}

export interface DriverDetails extends Driver {
  citizen?: CitizenInfo;
}

export interface Violation {
  id: number;
  driverId: number;
  type: string;
  description: string;
  date: string;
  location: string;
  points: number;
  fineAmount: number;
  status: 'active' | 'resolved';
}

export interface Fine {
  id: number;
  violationId: number;
  amount: number;
  paymentDeadline: string;
  paid: boolean;
  paymentDate: string | null;
}

export interface FineDetails extends Fine {
  violationType: string;
  violationDate: string;
  location: string;
  driverId: number;
  citizenId: number;
}

export interface Payment {
  id: number;
  fineId: number;
  citizenId: number;
  amount: number;
  status: 'pending' | 'completed';
  createdAt: string;
  completedAt: string | null;
}

export interface AppNotification {
  id: number;
  citizenId: number;
  message: string;
  type: string;
  read: boolean;
  createdAt: string;
}

export interface ViolationType {
  code: string;
  label: string;
  points: number;
  fine: number;
}

export interface ViolationStat {
  type: string;
  count: number;
  totalPoints: number;
  avgFine: number;
  totalFines: number;
}
