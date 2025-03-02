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
import { AuthService } from "../../../core/auth/auth.service";

@Component({
  selector: "app-sign-in",
  templateUrl: "./sign-in.component.html",
  styleUrls: ["./sign-in.component.css"],
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
export class SignInComponent {
  private authService = inject(AuthService);
  private router = inject(Router);
  
  signInForm: FormGroup;
  submitted = false;
  isLoading = false;
  error: string | null = null;
  
  focusedFields: { [key: string]: boolean } = {
    identifier: false,
    password: false,
  };

  constructor(private fb: FormBuilder) {
    this.signInForm = this.fb.group({
      identifier: ["", [Validators.required]],
      password: ["", [Validators.required]],
    });
  }

  showError(fieldName: string): boolean {
    const field = this.signInForm.get(fieldName);
    return (
      (field?.invalid && field?.touched && !this.focusedFields[fieldName]) || false
    );
  }

  setFieldFocus(isFocused: boolean, fieldName: string): void {
    this.focusedFields[fieldName] = isFocused;
    if (isFocused) {
      const field = this.signInForm.get(fieldName);
      if (field) {
        field.markAsUntouched();
      }
    }
  }

  onSubmit() {
    this.submitted = true;

    if (this.signInForm.valid) {
      this.isLoading = true;
      
      this.authService.login(this.signInForm.value).subscribe({
        next: () => {
          this.isLoading = false;
          toast.success('Successfully signed in');
          this.router.navigate(['/dashboard']);
        },
        error: (error) => {
          this.isLoading = false;
          toast.error(error.error?.message || 'Sign in failed. Please try again.');
          
          if (error.status === 401) {
            toast.error('Invalid credentials');
          }
        }
      });
    } else {
      Object.keys(this.signInForm.controls).forEach((key) => {
        const control = this.signInForm.get(key);
        if (control?.invalid) {
          control.markAsTouched();
        }
      });
    }
  }
}
