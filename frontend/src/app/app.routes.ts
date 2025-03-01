import { Routes } from '@angular/router';
import { SignUpComponent } from './features/auth/sign-up/sign-up.component';
import { LandingComponent } from './features/landing/landing.component';

export const routes: Routes = [
  { path: 'sign-up', component: SignUpComponent },
  { path: '', component: LandingComponent }
];
