import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from './auth.service';

export const authGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  return auth.isLoggedIn() ? true : inject(Router).createUrlTree(['/login']);
};

export const officerGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  return auth.isOfficer() ? true : inject(Router).createUrlTree(['/login']);
};

export const citizenGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  return auth.isCitizen() ? true : inject(Router).createUrlTree(['/login']);
};

export const adminGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  return auth.isAdmin() ? true : inject(Router).createUrlTree(['/login']);
};
