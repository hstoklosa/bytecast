import { Injectable, inject, signal, effect } from "@angular/core";
import { DOCUMENT } from "@angular/common";

export type Theme = "light" | "dark" | "system";

@Injectable({
  providedIn: "root",
})
export class ThemeService {
  private document = inject(DOCUMENT);
  private storageKey = "bytecast-theme";
  private defaultTheme: Theme = "system";

  // Private signal for internal state management
  private _theme = signal<Theme>(this.getInitialTheme());

  // Public readonly signal for consumers
  readonly theme = this._theme.asReadonly();

  // Media query for detecting system preference
  private systemPrefersDark = window.matchMedia("(prefers-color-scheme: dark)");

  constructor() {
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
  }

  /**
   * Sets the theme and persists it to localStorage
   */
  setTheme(newTheme: Theme): void {
    localStorage.setItem(this.storageKey, newTheme);
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
    const storedTheme = localStorage.getItem(this.storageKey) as Theme | null;
    return storedTheme || this.defaultTheme;
  }

  /**
   * Gets the system theme preference
   */
  private getSystemTheme(): "light" | "dark" {
    return this.systemPrefersDark.matches ? "dark" : "light";
  }

  /**
   * Applies the theme to the document element
   */
  private applyTheme(theme: Theme): void {
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
}
