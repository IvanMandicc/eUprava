import { Routes } from '@angular/router';
import { adminGuard, authGuard, citizenGuard, officerGuard } from './core/guards';
import { LoginComponent } from './pages/login.component';
import { RegisterComponent } from './pages/register.component';
import { CitizenDashboardComponent } from './pages/citizen/citizen-dashboard.component';
import { NotificationsComponent } from './pages/citizen/notifications.component';
import { DriversComponent } from './pages/officer/drivers.component';
import { ViolationsComponent } from './pages/officer/violations.component';
import { FinesComponent } from './pages/officer/fines.component';
import { UsersComponent } from './pages/admin/users.component';
import { OpenDataComponent } from './pages/open-data.component';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'login' },
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },

  // Open data — javna stranica, bez guard-a
  { path: 'open-data', component: OpenDataComponent },

  // Građanin
  { path: 'citizen', component: CitizenDashboardComponent, canActivate: [citizenGuard] },
  { path: 'citizen/notifications', component: NotificationsComponent, canActivate: [authGuard] },

  // Policajac
  { path: 'officer/drivers', component: DriversComponent, canActivate: [officerGuard] },
  { path: 'officer/violations', component: ViolationsComponent, canActivate: [officerGuard] },
  { path: 'officer/fines', component: FinesComponent, canActivate: [officerGuard] },

  // Administrator
  { path: 'admin/users', component: UsersComponent, canActivate: [adminGuard] },

  { path: '**', redirectTo: 'login' },
];
