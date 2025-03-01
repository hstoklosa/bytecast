import { Component } from "@angular/core";
import { CommonModule } from "@angular/common";
import {
  ReactiveFormsModule,
  FormGroup,
  FormBuilder,
  Validators,
} from "@angular/forms";
import { RouterLink } from "@angular/router";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import { HlmFormFieldComponent } from "@spartan-ng/ui-formfield-helm";
import { HlmInputDirective } from "@spartan-ng/ui-input-helm";
import { HlmLabelDirective } from "@spartan-ng/ui-label-helm";
import { HlmErrorDirective } from "@spartan-ng/ui-formfield-helm";

@Component({
  selector: "app-sign-up",
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    HlmButtonDirective,
    HlmFormFieldComponent,
    HlmInputDirective,
    HlmLabelDirective,
    HlmErrorDirective,
  ],
  template: `
    <div
      class="flex min-h-svh flex-col items-center justify-center gap-6 bg-background p-2 md:p-6"
    >
      <div class="w-full max-w-[450px]">
        <div class="flex flex-col gap-3">
          <div class="flex flex-col items-center gap-2">
            <a
              href="#"
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
            <h1 class="text-xl font-bold">Welcome to Bytecast</h1>
            <div class="text-center text-sm">
              Already have an account?
              <a
                routerLink="/login"
                class="underline underline-offset-4"
                >Sign in</a
              >
            </div>
          </div>

          <div class="p-6 rounded-lg">
            <form
              [formGroup]="signUpForm"
              (ngSubmit)="onSubmit()"
              class="grid gap-4"
            >
              <!-- Username Field -->
              <hlm-form-field class="grid gap-1">
                <label
                  hlmLabel
                  for="username"
                  >Username</label
                >
                <input
                  hlmInput
                  id="username"
                  type="text"
                  formControlName="username"
                  [class.border-destructive]="showError('username')"
                  placeholder="Enter username"
                  class="w-full"
                />
                <div
                  *ngIf="showError('username') && submitted"
                  class="text-destructive text-sm"
                  hlmError
                >
                  <span *ngIf="signUpForm.get('username')?.errors?.['required']">
                    Username is required
                  </span>
                  <span *ngIf="signUpForm.get('username')?.errors?.['minlength']">
                    Username must be at least 3 characters
                  </span>
                </div>
              </hlm-form-field>

              <!-- Email Field -->
              <hlm-form-field class="grid gap-1">
                <label
                  hlmLabel
                  for="email"
                  >Email</label
                >
                <input
                  hlmInput
                  id="email"
                  type="email"
                  formControlName="email"
                  [class.border-destructive]="showError('email')"
                  placeholder="name@example.com"
                  class="w-full"
                />
                <div
                  *ngIf="showError('email') && submitted"
                  class="text-destructive text-sm"
                  hlmError
                >
                  <span *ngIf="signUpForm.get('email')?.errors?.['required']">
                    Email is required
                  </span>
                  <span *ngIf="signUpForm.get('email')?.errors?.['email']">
                    Please enter a valid email
                  </span>
                </div>
              </hlm-form-field>

              <!-- Password Field -->
              <hlm-form-field class="grid gap-1">
                <label
                  hlmLabel
                  for="password"
                  >Password</label
                >
                <input
                  hlmInput
                  id="password"
                  type="password"
                  formControlName="password"
                  [class.border-destructive]="showError('password')"
                  placeholder="Enter password"
                  class="w-full"
                />
                <div
                  *ngIf="showError('password') && submitted"
                  class="text-destructive text-sm"
                  hlmError
                >
                  <span *ngIf="signUpForm.get('password')?.errors?.['required']">
                    Password is required
                  </span>
                  <span *ngIf="signUpForm.get('password')?.errors?.['minlength']">
                    Password must be at least 8 characters
                  </span>
                </div>
              </hlm-form-field>

              <!-- Confirm Password Field -->
              <hlm-form-field class="grid gap-1">
                <label
                  hlmLabel
                  for="confirmPassword"
                  >Confirm Password</label
                >
                <input
                  hlmInput
                  id="confirmPassword"
                  type="password"
                  formControlName="confirmPassword"
                  [class.border-destructive]="showError('confirmPassword')"
                  placeholder="Confirm password"
                  class="w-full"
                />
                <div
                  *ngIf="showError('confirmPassword') && submitted"
                  class="text-destructive text-sm"
                  hlmError
                >
                  <span
                    *ngIf="signUpForm.get('confirmPassword')?.errors?.['required']"
                  >
                    Please confirm your password
                  </span>
                  <span
                    *ngIf="signUpForm.get('confirmPassword')?.errors?.['passwordMismatch']"
                  >
                    Passwords do not match
                  </span>
                </div>
              </hlm-form-field>

              <button
                hlmBtn
                type="submit"
                [disabled]="signUpForm.invalid"
                class="w-full"
              >
                Create account
              </button>
            </form>

            <div
              class="relative text-center text-sm after:absolute after:inset-0 after:top-1/2 after:z-0 after:flex after:items-center after:border-t after:border-border mt-6"
            >
              <span class="relative z-10 bg-background px-2 text-muted-foreground"
                >OR</span
              >
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-6">
              <button
                hlmBtn
                variant="outline"
                type="button"
                class="w-full cursor-pointer"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  class="size-5 mr-2"
                >
                  <path
                    d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09z"
                    fill="currentColor"
                  />
                  <path
                    d="M15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701"
                    fill="currentColor"
                  />
                </svg>
              </button>
              <button
                hlmBtn
                variant="outline"
                type="button"
                class="w-full cursor-pointer"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  class="size-5 mr-2"
                >
                  <path
                    d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z"
                    fill="currentColor"
                  />
                </svg>
              </button>
            </div>
          </div>

          <div
            class="text-balance text-center text-xs text-muted-foreground [&_a]:underline [&_a]:underline-offset-4 hover:[&_a]:text-primary"
          >
            By clicking continue, you agree to our
            <a href="#">Terms of Service</a> and <a href="#">Privacy Policy</a>.
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [
    `
      :host {
        display: block;
        background: hsl(var(--background));
      }

      .hlm-input {
        transition: all 200ms ease;
      }

      .hlm-input:focus {
        outline: none;
        box-shadow: 0 0 0 2px hsl(var(--background)), 0 0 0 4px hsl(var(--ring));
      }

      .hlm-error {
        animation: shake 0.2s ease-in-out 0s 2;
        color: hsl(var(--destructive));
        font-size: 0.875rem;
        margin-top: 0.375rem;
      }

      @keyframes shake {
        0%,
        100% {
          transform: translateX(0);
        }
        25% {
          transform: translateX(-2px);
        }
        75% {
          transform: translateX(2px);
        }
      }
    `,
  ],
})
export class SignUpComponent {
  signUpForm: FormGroup;
  submitted = false;

  constructor(private fb: FormBuilder) {
    this.signUpForm = this.fb.group(
      {
        username: ["", [Validators.required, Validators.minLength(3)]],
        email: ["", [Validators.required, Validators.email]],
        password: ["", [Validators.required, Validators.minLength(8)]],
        confirmPassword: ["", [Validators.required]],
      },
      {
        validators: this.passwordMatchValidator,
      }
    );
  }

  passwordMatchValidator(form: FormGroup) {
    const password = form.get("password");
    const confirmPassword = form.get("confirmPassword");

    if (password && confirmPassword && password.value !== confirmPassword.value) {
      confirmPassword.setErrors({ passwordMismatch: true });
    } else {
      confirmPassword?.setErrors(null);
    }
  }

  showError(fieldName: string): boolean {
    const field = this.signUpForm.get(fieldName);
    return field ? field.invalid && (field.dirty || field.touched) : false;
  }

  onSubmit() {
    this.submitted = true;
    if (this.signUpForm.valid) {
      console.log("Form submitted:", this.signUpForm.value);
      // TODO: Implement sign-up logic
    } else {
      Object.keys(this.signUpForm.controls).forEach((key) => {
        const control = this.signUpForm.get(key);
        if (control?.invalid) {
          control.markAsTouched();
        }
      });
    }
  }
}
