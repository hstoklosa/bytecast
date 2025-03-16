import { inject } from "@angular/core";
import { Router, type CanActivateFn } from "@angular/router";
import { map } from "rxjs";

import { AuthService } from "../services";

export const authGuard: CanActivateFn = (route, state) => {
  const router = inject(Router);
  const authService = inject(AuthService);

  return authService.isAuthenticated$.pipe(
    map((isAuthenticated) => {
      if (isAuthenticated) {
        return true;
      }

      // Store attempted URL for redirection after login
      router.navigate(["/"], { queryParams: { returnUrl: state.url } });
      return false;
    })
  );
};
