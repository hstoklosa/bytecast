import { ChangeDetectionStrategy, Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import { RouterLink } from "@angular/router";

@Component({
  selector: "app-auth-layout",
  standalone: true,
  imports: [CommonModule, RouterLink],
  template: `
    <div
      class="flex min-h-svh flex-col items-center justify-center gap-6 bg-background p-2 md:p-6"
    >
      <div class="w-full max-w-[450px]">
        <div class="flex flex-col gap-3">
          <div class="flex flex-col items-center gap-2">
            <a
              routerLink="/"
              class="flex flex-col items-center gap-2 font-medium"
            >
              <div
                class="flex h-8 w-8 items-center justify-center rounded-md text-primary"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="lucide lucide-gallery-vertical-end size-6"
                >
                  <path d="M7 2h10"></path>
                  <path d="M5 6h14"></path>
                  <rect
                    width="18"
                    height="12"
                    x="3"
                    y="10"
                    rx="2"
                  ></rect>
                </svg>
              </div>
              <span class="sr-only">Bytecast</span>
            </a>

            <!-- Content projection for title and links -->
            <ng-content select="[auth-header]"></ng-content>
          </div>

          <div class="p-6 rounded-lg">
            <!-- Content projection for form -->
            <ng-content select="[auth-form]"></ng-content>
          </div>

          <div
            class="text-balance text-center text-xs text-muted-foreground [&_a]:underline [&_a]:underline-offset-4 hover:[&_a]:text-primary"
          >
            By continuing, you agree to our
            <a href="#">Terms of Service</a> and <a href="#">Privacy Policy</a>.
          </div>
        </div>
      </div>
    </div>
  `,
  styleUrls: ["./auth-layout.component.css"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class AuthLayoutComponent {}
