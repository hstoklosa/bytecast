import { Routes } from '@angular/router';
import { authGuard } from './core/auth/auth.guard';
import { guestGuard } from './core/auth/guest.guard';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./features/landing/landing.component').then(m => m.LandingComponent),
    canActivate: [guestGuard]
  },
  {
    path: 'sign-in',
    loadComponent: () =>
      import('./features/auth/sign-in/sign-in.component').then(m => m.SignInComponent),
    canActivate: [guestGuard]
  },
  {
    path: 'sign-up',
    loadComponent: () =>
      import('./features/auth/sign-up/sign-up.component').then(m => m.SignUpComponent),
    canActivate: [guestGuard]
  },
  {
    path: 'dashboard',
    loadComponent: () =>
      import('./features/dashboard/dashboard.component').then(m => m.DashboardComponent),
    canActivate: [authGuard]
  }
];
