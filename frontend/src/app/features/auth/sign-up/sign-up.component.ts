import { Component, inject } from "@angular/core";
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
import { AuthService } from "../../../core/auth/auth.service";

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
        password: ["", [
          Validators.required,
          Validators.minLength(8),
          Validators.pattern(/^(?=.*[a-zA-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&].*$/),
        ]],
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
          toast.success('Account created successfully');
          this.router.navigate(['/dashboard']);
        },
        error: (error) => {
          this.isLoading = false;
          toast.error(error.error?.message || 'Registration failed. Please try again.');
          
          if (error.status === 409) {
            if (error.error?.error?.includes('Email')) {
              this.signUpForm.get('email')?.setErrors({ exists: true });
              toast.error('Email already exists');
            } else if (error.error?.error?.includes('Username')) {
              this.signUpForm.get('username')?.setErrors({ taken: true });
              toast.error('Username already taken');
            }
          }
        }
      });
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
