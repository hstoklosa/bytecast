import { Component, inject } from "@angular/core";
import { RouterOutlet } from "@angular/router";
import { HlmToasterComponent } from "@spartan-ng/ui-sonner-helm";
import { ThemeService } from "./core/services/theme.service";

@Component({
  selector: "app-root",
  standalone: true,
  imports: [RouterOutlet, HlmToasterComponent],
  template: `
    <hlm-toaster />
    <router-outlet></router-outlet>
  `,
})
export class AppComponent {
  // Inject ThemeService to ensure it's initialized
  private themeService = inject(ThemeService);
}
