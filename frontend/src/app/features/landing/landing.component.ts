import { Component, OnDestroy, OnInit } from "@angular/core";
import { CommonModule } from "@angular/common";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import { Router } from "@angular/router";
import { Feature } from "./landing.interface";
import { LucideAngularModule, MoveRight, Info } from "lucide-angular";
import {
  animate,
  keyframes,
  style,
  transition,
  trigger,
} from "@angular/animations";

@Component({
  selector: "app-landing",
  standalone: true,
  imports: [CommonModule, HlmButtonDirective, LucideAngularModule],
  animations: [
    trigger("titleAnimation", [
      transition(":enter, * => *", [
        animate(
          "400ms cubic-bezier(0.4, 0, 0.2, 1)",
          keyframes([
            style({ opacity: 0, transform: "translateY(100%)", offset: 0 }),
            style({ opacity: 1, transform: "translateY(0)", offset: 1 }),
          ])
        ),
      ]),
    ]),
  ],
  templateUrl: "./landing.component.html",
  styleUrls: ["./landing.component.css"],
})
export class LandingComponent implements OnInit, OnDestroy {
  readonly moveRightIcon = MoveRight;
  readonly infoIcon = Info;
  private intervalId?: number;
  titleNumber = 0;
  readonly titles = [
    "automated",
    "personalized",
    "summarized",
    "simplified",
    "organized",
  ];
  // Commenting out features for now
  private readonly features: Feature[] = [
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

  ngOnInit(): void {
    this.startTitleAnimation();
  }

  ngOnDestroy(): void {
    if (this.intervalId) {
      window.clearInterval(this.intervalId);
    }
  }

  private startTitleAnimation(): void {
    this.intervalId = window.setInterval(() => {
      this.titleNumber = (this.titleNumber + 1) % this.titles.length;
    }, 2000);
  }

  navigateToSignup(): void {
    this.router.navigate(["/sign-up"]);
  }
}
