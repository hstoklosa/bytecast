import { Component, inject } from "@angular/core";
import { HttpErrorResponse } from "@angular/common/http";
import { CommonModule } from "@angular/common";
import {
  ReactiveFormsModule,
  FormGroup,
  FormBuilder,
  Validators,
} from "@angular/forms";
import { Router, RouterLink } from "@angular/router";
import { toast } from "ngx-sonner";
import { HlmButtonDirective } from "@spartan-ng/ui-button-helm";
import { HlmFormFieldComponent } from "@spartan-ng/ui-formfield-helm";
import { HlmInputDirective } from "@spartan-ng/ui-input-helm";
import { HlmLabelDirective } from "@spartan-ng/ui-label-helm";
import { HlmSpinnerComponent } from "@spartan-ng/ui-spinner-helm";
import { HlmToasterComponent } from "@spartan-ng/ui-sonner-helm";

import { passwordMatchValidator } from "./sign-up.validators";

import { AuthService } from "../../../core/services";
import { AuthLayoutComponent } from "../../../layout";

@Component({
  selector: "app-sign-up",
  templateUrl: "./sign-up.component.html",
  styleUrls: ["./sign-up.component.css"],
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    RouterLink,
    HlmButtonDirective,
    HlmFormFieldComponent,
    HlmInputDirective,
    HlmLabelDirective,
    HlmSpinnerComponent,
    HlmToasterComponent,
    AuthLayoutComponent,
  ],
})
export class SignUpComponent {
  private authService = inject(AuthService);
  private router = inject(Router);

  signUpForm: FormGroup;
  submitted = false;
  isLoading = false;
  error: string | null = null;

  focusedFields: { [key: string]: boolean } = {
    username: false,
    email: false,
    password: false,
    confirmPassword: false,
  };

  constructor(private fb: FormBuilder) {
    this.signUpForm = this.fb.group(
      {
        username: ["", [Validators.required, Validators.minLength(3)]],
        email: ["", [Validators.required, Validators.email]],
        password: [
          "",
          [
            Validators.required,
            Validators.minLength(8),
            Validators.pattern(
              /^(?=.*[a-zA-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&].*$/
            ),
          ],
        ],
        confirmPassword: ["", [Validators.required]],
      },
      {
        validators: passwordMatchValidator,
      }
    );
  }

  showError(fieldName: string): boolean {
    const field = this.signUpForm.get(fieldName);
    return (
      (field?.invalid && field?.touched && !this.focusedFields[fieldName]) || false
    );
  }

  setFieldFocus(isFocused: boolean, fieldName: string): void {
    this.focusedFields[fieldName] = isFocused;
    if (isFocused) {
      const field = this.signUpForm.get(fieldName);
      if (field) {
        field.markAsUntouched();
      }
    }
  }

  onSubmit() {
    this.submitted = true;

    if (this.signUpForm.valid) {
      this.isLoading = true;

      const { confirmPassword, ...registrationData } = this.signUpForm.value;

      this.authService.register(registrationData).subscribe({
        next: () => {
          this.isLoading = false;
          toast.success("Account created successfully");
          this.router.navigate(["/dashboard"]);
        },
        error: (error) => {
          this.isLoading = false;
          toast.error(error.message || "Registration failed. Please try again.");

          const message = error.message?.toLowerCase() || "";
          if (message.includes("email")) {
            this.signUpForm.get("email")?.setErrors({ exists: true });
          } else if (message.includes("username")) {
            this.signUpForm.get("username")?.setErrors({ taken: true });
          }
        },
      });
    } else {
      // Show toast for invalid form
      let errorMessage = "Please fill all required fields correctly";

      // Check for specific validation errors to provide more helpful messages
      if (this.signUpForm.hasError("passwordMismatch")) {
        errorMessage = "Passwords do not match";
      } else if (this.signUpForm.get("password")?.hasError("pattern")) {
        errorMessage =
          "Password must contain at least one letter, one number, and one special character";
      } else if (this.signUpForm.get("email")?.hasError("email")) {
        errorMessage = "Please enter a valid email address";
      }

      toast.error(errorMessage);

      // Mark invalid fields as touched to show validation errors
      Object.keys(this.signUpForm.controls).forEach((key) => {
        const control = this.signUpForm.get(key);
        if (control?.invalid) {
          control.markAsTouched();
        }
      });
    }
  }
}
