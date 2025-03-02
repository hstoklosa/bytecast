import { Component } from "@angular/core";
import { RouterOutlet } from "@angular/router";
import { HlmToasterComponent } from "@spartan-ng/ui-sonner-helm";

@Component({
  selector: "app-root",
  standalone: true,
  imports: [RouterOutlet, HlmToasterComponent],
  template: `
    <hlm-toaster />
    <router-outlet></router-outlet>
  `,
})
export class AppComponent {}
