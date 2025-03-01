import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import { Router } from "@angular/router";

interface Feature {
  title: string;
  description: string;
  icon: string;
}

@Component({
  selector: "app-landing",
  standalone: true,
  imports: [CommonModule, HlmButtonDirective],
  template: `
    <div
      class="min-h-screen bg-gradient-to-b from-background to-card md:h-screen md:overflow-hidden"
    >
      <div class="container mx-auto px-4 py-8 md:py-0 max-w-6xl">
        <div
          class="md:min-h-screen md:grid md:grid-cols-2 md:gap-8 md:items-center"
        >
          <!-- Left Column: Hero Section -->
          <div class="text-center md:text-left mb-12 md:mb-0 md:pr-8">
            <h1
              class="text-4xl md:text-5xl font-bold text-primary mb-6"
            >
              ByteCast
            </h1>
            <p class="text-xl text-muted-foreground mb-8 max-w-xl">
              Transform your favorite YouTube channels into personalized podcast
              summaries
            </p>
            <button
              hlmBtn
              variant="default"
              size="lg"
              class="px-6 py-2.5 text-base"
              (click)="navigateToSignup()"
            >
              Create Your Watchlist
            </button>
          </div>

          <!-- Right Column: Features -->
          <div class="space-y-4">
            <div
              *ngFor="let feature of features"
              class="feature-card bg-card p-5 rounded-lg shadow-sm hover:shadow-md transition-shadow border"
            >
              <div class="flex items-start space-x-4">
                <div class="flex-none">
                  <div
                    class="w-10 h-10 bg-secondary rounded-lg flex items-center justify-center"
                  >
                    <svg
                      class="w-5 h-5 text-primary"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        [attr.d]="feature.icon"
                      />
                    </svg>
                  </div>
                </div>
                <div class="flex-1 min-w-0">
                  <h3 class="text-lg font-semibold text-primary mb-1">
                    {{ feature.title }}
                  </h3>
                  <p class="text-sm text-muted-foreground leading-relaxed">
                    {{ feature.description }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class LandingComponent {
  readonly features: Feature[] = [
    {
      title: "Curated Watchlists",
      description:
        "Add your favorite YouTube channels and organize them in personalized watchlists. Easy search and management.",
      icon: "M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2",
    },
    {
      title: "Auto Content Updates",
      description:
        "Never miss a video. We automatically detect and collect new uploads from your selected channels.",
      icon: "M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9",
    },
    {
      title: "Smart Summaries",
      description:
        "Get weekly podcast-style summaries of all new content, organized by channel and video uploads.",
      icon: "M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z",
    },
  ];

  constructor(private router: Router) {}

  navigateToSignup(): void {
    this.router.navigate(["/sign-up"]);
  }
}
