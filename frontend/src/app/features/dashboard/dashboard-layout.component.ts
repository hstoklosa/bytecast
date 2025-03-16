import { ChangeDetectionStrategy, Component, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import { BrnMenuTriggerDirective } from "@spartan-ng/brain/menu";
import { NgIconComponent } from "@ng-icons/core";
import {
  HlmMenuComponent,
  HlmMenuGroupComponent,
  HlmMenuItemDirective,
  HlmMenuItemIconDirective,
  HlmMenuLabelComponent,
  HlmMenuSeparatorComponent,
} from "@spartan-ng/ui-menu-helm";
import { provideIcons } from "@ng-icons/core";
import { lucideSettings, lucideUser } from "@ng-icons/lucide";
import { HlmIconDirective } from "@spartan-ng/ui-icon-helm";
import { AuthService } from "../../core/auth/auth.service";
import { ThemeToggleComponent } from "../../shared/components/theme-toggle/theme-toggle.component";

@Component({
  selector: "app-dashboard-layout",
  standalone: true,
  imports: [
    CommonModule,
    HlmButtonDirective,
    BrnMenuTriggerDirective,
    HlmMenuComponent,
    HlmMenuGroupComponent,
    HlmMenuItemDirective,
    HlmMenuItemIconDirective,
    HlmMenuLabelComponent,
    HlmMenuSeparatorComponent,
    HlmIconDirective,
    NgIconComponent,
    ThemeToggleComponent,
  ],
  providers: [provideIcons({ lucideUser, lucideSettings })],
  template: `
    <div class="min-h-screen bg-background py-6 space-y-8">
      <header class="sticky top-0 z-50 w-full bg-background/95 backdrop-blur">
        <div class="container flex h-14 items-center justify-between">
          <h1 class="text-3xl font-bold tracking-tight">Bytecast</h1>
          <div class="flex items-center gap-2">
            <app-theme-toggle />
            <button
              hlmBtn
              variant="ghost"
              size="icon"
              align="end"
              [brnMenuTriggerFor]="menu"
            >
              <ng-icon
                hlm
                name="lucideUser"
              />
            </button>
          </div>

          <ng-template #menu>
            <hlm-menu class="w-36">
              <hlm-menu-group>
                <button hlmMenuItem>
                  <ng-icon
                    hlm
                    name="lucideSettings"
                    class="mr-2"
                    size="16px"
                  />
                  <span>Settings</span>
                </button>

                <hlm-menu-separator />

                <button
                  hlmMenuItem
                  (click)="onLogout()"
                >
                  Logout
                </button>
              </hlm-menu-group>
            </hlm-menu>
          </ng-template>
        </div>
      </header>

      <main class="container">
        <ng-content></ng-content>
      </main>
    </div>
  `,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DashboardLayoutComponent {
  private authService = inject(AuthService);

  onLogout(): void {
    this.authService.logout().subscribe();
  }
}
