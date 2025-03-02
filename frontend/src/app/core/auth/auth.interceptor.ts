import { inject } from '@angular/core';
import {
  HttpRequest,
  HttpHandlerFn,
  HttpInterceptorFn,
  HttpErrorResponse,
} from '@angular/common/http';
import { Observable, catchError, switchMap, throwError } from 'rxjs';
import { AuthService } from './auth.service';

let isRefreshing = false;

export const authInterceptor: HttpInterceptorFn = (
  request: HttpRequest<unknown>,
  next: HttpHandlerFn
) => {
  const authService = inject(AuthService);

  const addToken = (req: HttpRequest<unknown>): HttpRequest<unknown> => {
    const token = authService.getToken();
    if (token) {
      return req.clone({
        setHeaders: {
          Authorization: `Bearer ${token}`,
        },
      });
    }
    return req;
  };

  const handle401Error = (
    req: HttpRequest<unknown>,
    handler: HttpHandlerFn
  ): Observable<any> => {
    if (!isRefreshing) {
      isRefreshing = true;

      return authService.refreshToken().pipe(
        switchMap(() => {
          isRefreshing = false;
          return handler(addToken(req));
        }),
        catchError((error) => {
          isRefreshing = false;
          return throwError(() => error);
        })
      );
    }

    return handler(req);
  };

  const isAuthEndpoint = (url: string): boolean =>
    url.includes('/api/v1/auth/');

  const isRefreshRequest = (url: string): boolean =>
    url.includes('/api/v1/auth/refresh');

  // Skip adding token for auth endpoints (except refresh)
  if (isAuthEndpoint(request.url) && !isRefreshRequest(request.url)) {
    return next(request);
  }

  return next(addToken(request)).pipe(
    catchError((error: HttpErrorResponse) => {
      if (error.status === 401 && !isRefreshRequest(request.url)) {
        return handle401Error(request, next);
      }
      return throwError(() => error);
    })
  );
};
