import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, catchError, map, of, tap } from 'rxjs';
import { Router } from '@angular/router';

export interface AuthResponse {
  access_token: string;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private http = inject(HttpClient);
  private router = inject(Router);
  
  private readonly AUTH_TOKEN_KEY = 'access_token';
  private readonly API_URL = '/api/v1/auth';
  
  private isAuthenticatedSubject = new BehaviorSubject<boolean>(this.hasValidToken());
  isAuthenticated$ = this.isAuthenticatedSubject.asObservable();

  constructor() {
    // Check token validity on service initialization
    this.isAuthenticatedSubject.next(this.hasValidToken());
  }

  register(data: RegisterRequest): Observable<void> {
    return this.http.post<AuthResponse>(`${this.API_URL}/register`, data).pipe(
      tap(response => {
        this.setToken(response.access_token);
        this.isAuthenticatedSubject.next(true);
      }),
      map(() => void 0),
      catchError(error => {
        console.error('Registration failed:', error);
        throw error;
      })
    );
  }

  login(data: LoginRequest): Observable<void> {
    return this.http.post<AuthResponse>(`${this.API_URL}/login`, data).pipe(
      tap(response => {
        this.setToken(response.access_token);
        this.isAuthenticatedSubject.next(true);
      }),
      map(() => void 0),
      catchError(error => {
        console.error('Login failed:', error);
        throw error;
      })
    );
  }

  logout(): Observable<void> {
    return this.http.post<void>(`${this.API_URL}/logout`, {}).pipe(
      tap(() => {
        this.clearToken();
        this.isAuthenticatedSubject.next(false);
        this.router.navigate(['/']);
      }),
      catchError(error => {
        console.error('Logout failed:', error);
        // Still clear token and redirect even if logout fails
        this.clearToken();
        this.isAuthenticatedSubject.next(false);
        this.router.navigate(['/']);
        return of(void 0);
      })
    );
  }

  refreshToken(): Observable<void> {
    return this.http.post<AuthResponse>(`${this.API_URL}/refresh`, {}).pipe(
      tap(response => {
        this.setToken(response.access_token);
        this.isAuthenticatedSubject.next(true);
      }),
      map(() => void 0),
      catchError(error => {
        console.error('Token refresh failed:', error);
        if (error.status === 401) {
          this.clearToken();
          this.isAuthenticatedSubject.next(false);
          this.router.navigate(['/']);
        }
        throw error;
      })
    );
  }

  private setToken(token: string): void {
    localStorage.setItem(this.AUTH_TOKEN_KEY, token);
  }

  private clearToken(): void {
    localStorage.removeItem(this.AUTH_TOKEN_KEY);
  }

  getToken(): string | null {
    return localStorage.getItem(this.AUTH_TOKEN_KEY);
  }

  private hasValidToken(): boolean {
    const token = this.getToken();
    if (!token) return false;

    try {
      // Simple structural validation - check if it's a valid JWT format
      const [header, payload, signature] = token.split('.');
      if (!header || !payload || !signature) return false;

      // Check expiration
      const decodedPayload = JSON.parse(atob(payload));
      if (!decodedPayload.exp) return false;

      const expirationDate = new Date(decodedPayload.exp * 1000);
      return expirationDate > new Date();
    } catch {
      return false;
    }
  }
}
