import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../core/auth/auth.service';
import { HlmButtonDirective } from '@spartan-ng/ui-button-helm';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [RouterLink, HlmButtonDirective],
  template: `
    <div class="container mx-auto p-4">
      <h1 class="text-2xl font-bold mb-4">Dashboard</h1>
      <p class="mb-4">Welcome to your dashboard!</p>
      <button hlmBtn type="button" (click)="onLogout()">Logout</button>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DashboardComponent {
  private authService = inject(AuthService);

  onLogout(): void {
    this.authService.logout().subscribe();
  }
}
