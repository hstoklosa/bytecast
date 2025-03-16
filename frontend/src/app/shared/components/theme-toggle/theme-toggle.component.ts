import { Component, inject } from "@angular/core";
import { CommonModule } from "@angular/common";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import { NgIconComponent, provideIcons } from "@ng-icons/core";
import { HlmIconDirective } from "@spartan-ng/ui-icon-helm";
import { lucideMoon, lucideSun } from "@ng-icons/lucide";

import { Theme } from "../../../core/models";
import { ThemeService } from "../../../core/services";

@Component({
  selector: "app-theme-toggle",
  standalone: true,
  imports: [CommonModule, HlmButtonDirective, HlmIconDirective, NgIconComponent],
  providers: [provideIcons({ lucideSun, lucideMoon })],
  templateUrl: "./theme-toggle.component.html",
  // styleUrls: ["./theme-toggle.component.scss"],
})
export class ThemeToggleComponent {
  private themeService = inject(ThemeService);
  readonly currentTheme = this.themeService.theme;

  effectiveTheme = () => this.themeService.getEffectiveTheme();

  toggleTheme(): void {
    this.themeService.toggleTheme();
  }

  setTheme(theme: Theme): void {
    this.themeService.setTheme(theme);
  }
}
