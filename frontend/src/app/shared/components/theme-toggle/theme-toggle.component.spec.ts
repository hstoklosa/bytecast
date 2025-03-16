import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ThemeToggleComponent } from "./theme-toggle.component";
import { ThemeService } from "../../../core/services/theme.service";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import { NgIconComponent, provideIcons } from "@ng-icons/core";
import { HlmIconDirective } from "@spartan-ng/ui-icon-helm";
import { lucideMoon, lucideSun } from "@ng-icons/lucide";

describe("ThemeToggleComponent", () => {
  let component: ThemeToggleComponent;
  let fixture: ComponentFixture<ThemeToggleComponent>;
  let themeServiceSpy: jasmine.SpyObj<ThemeService>;

  beforeEach(async () => {
    // Create a spy for ThemeService
    const spy = jasmine.createSpyObj("ThemeService", [
      "getEffectiveTheme",
      "toggleTheme",
      "setTheme",
    ]);

    // Mock the theme signal
    Object.defineProperty(spy, "theme", {
      get: () => ({ asReadonly: () => () => "light" }),
    });

    await TestBed.configureTestingModule({
      imports: [
        ThemeToggleComponent,
        HlmButtonDirective,
        HlmIconDirective,
        NgIconComponent,
      ],
      providers: [
        provideIcons({ lucideSun, lucideMoon }),
        { provide: ThemeService, useValue: spy },
      ],
    }).compileComponents();

    themeServiceSpy = TestBed.inject(ThemeService) as jasmine.SpyObj<ThemeService>;
    themeServiceSpy.getEffectiveTheme.and.returnValue("light");
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(ThemeToggleComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("should call toggleTheme when button is clicked", () => {
    const button = fixture.nativeElement.querySelector("button");
    button.click();
    expect(themeServiceSpy.toggleTheme).toHaveBeenCalled();
  });

  it("should display the correct icon based on the current theme", () => {
    // Test light theme
    themeServiceSpy.getEffectiveTheme.and.returnValue("light");
    fixture.detectChanges();
    let icon = fixture.nativeElement.querySelector("ng-icon");
    expect(icon.getAttribute("ng-reflect-name")).toBe("lucideMoon");

    // Test dark theme
    themeServiceSpy.getEffectiveTheme.and.returnValue("dark");
    fixture.detectChanges();
    icon = fixture.nativeElement.querySelector("ng-icon");
    expect(icon.getAttribute("ng-reflect-name")).toBe("lucideSun");
  });

  it("should call setTheme with the correct theme", () => {
    component.setTheme("dark");
    expect(themeServiceSpy.setTheme).toHaveBeenCalledWith("dark");
  });
});
