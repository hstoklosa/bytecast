import {
  Injectable,
  inject,
  signal,
  effect,
  PLATFORM_ID,
  Inject,
} from "@angular/core";
import { DOCUMENT, isPlatformBrowser } from "@angular/common";

export type Theme = "light" | "dark" | "system";

@Injectable({
  providedIn: "root",
})
export class ThemeService {
  private document = inject(DOCUMENT);
  private isBrowser: boolean;
  private storageKey = "bytecast-theme";
  private defaultTheme: Theme = "system";
  private transitionStyleElement: HTMLStyleElement | null = null;

  // Private signal for internal state management
  private _theme = signal<Theme>(this.getInitialTheme());

  // Public readonly signal for consumers
  readonly theme = this._theme.asReadonly();

  // Media query for detecting system preference
  private systemPrefersDark: MediaQueryList | null = null;

  constructor(@Inject(PLATFORM_ID) platformId: Object) {
    this.isBrowser = isPlatformBrowser(platformId);

    if (this.isBrowser) {
      this.systemPrefersDark = window.matchMedia("(prefers-color-scheme: dark)");

      // Set up effect to apply theme changes
      effect(() => {
        this.applyTheme(this._theme());
      });

      // Listen for system preference changes
      this.systemPrefersDark.addEventListener("change", () => {
        if (this._theme() === "system") {
          this.applyTheme("system");
        }
      });

      // Add global style to disable transitions during theme changes
      this.addTransitionDisablingStyle();
    }
  }

  /**
   * Sets the theme and persists it to localStorage
   */
  setTheme(newTheme: Theme): void {
    if (this.isBrowser) {
      this.disableTransitions();
      localStorage.setItem(this.storageKey, newTheme);
    }
    this._theme.set(newTheme);
  }

  /**
   * Toggles between light and dark themes
   * If current theme is system, it will switch to either light or dark based on current system preference
   */
  toggleTheme(): void {
    const currentTheme = this._theme();

    if (currentTheme === "system") {
      // If system, toggle to the opposite of system preference
      this.setTheme(this.getSystemTheme() === "dark" ? "light" : "dark");
    } else {
      // Toggle between light and dark
      this.setTheme(currentTheme === "dark" ? "light" : "dark");
    }
  }

  /**
   * Gets the current effective theme (resolves 'system' to either 'light' or 'dark')
   */
  getEffectiveTheme(): "light" | "dark" {
    const currentTheme = this._theme();
    return currentTheme === "system" ? this.getSystemTheme() : currentTheme;
  }

  /**
   * Retrieves the initial theme from localStorage or falls back to default
   */
  private getInitialTheme(): Theme {
    if (!this.isBrowser) {
      return this.defaultTheme;
    }

    const storedTheme = localStorage.getItem(this.storageKey) as Theme | null;
    return storedTheme || this.defaultTheme;
  }

  /**
   * Gets the system theme preference
   */
  private getSystemTheme(): "light" | "dark" {
    if (!this.isBrowser || !this.systemPrefersDark) {
      return "light";
    }
    return this.systemPrefersDark.matches ? "dark" : "light";
  }

  /**
   * Applies the theme to the document element
   */
  private applyTheme(theme: Theme): void {
    if (!this.isBrowser) {
      return;
    }

    const root = this.document.documentElement;

    // Remove both theme classes
    root.classList.remove("light", "dark");

    // Apply the appropriate theme
    if (theme === "system") {
      root.classList.add(this.getSystemTheme());
    } else {
      root.classList.add(theme);
    }
  }

  /**
   * Temporarily disables all transitions in the app during theme changes
   */
  private disableTransitions(): void {
    if (!this.isBrowser) {
      return;
    }

    const root = this.document.documentElement;
    root.classList.add("disable-transitions");

    // Use requestAnimationFrame for better timing with browser rendering
    requestAnimationFrame(() => {
      // Re-enable transitions after a short delay
      setTimeout(() => {
        root.classList.remove("disable-transitions");
      }, 100);
    });
  }

  /**
   * Adds a style element to the document head that defines the disable-transitions class
   */
  private addTransitionDisablingStyle(): void {
    if (!this.isBrowser || this.transitionStyleElement) {
      return;
    }

    this.transitionStyleElement = this.document.createElement("style");
    this.transitionStyleElement.setAttribute("id", "theme-transition-style");
    this.transitionStyleElement.textContent = `
      .disable-transitions,
      .disable-transitions *,
      .disable-transitions *::before,
      .disable-transitions *::after {
        transition: none !important;
        animation: none !important;
      }
    `;
    this.document.head.appendChild(this.transitionStyleElement);
  }
}
