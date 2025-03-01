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
import { passwordMatchValidator } from "./sign-up.validators";

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
    HlmErrorDirective,
  ],
})
export class SignUpComponent {
  signUpForm: FormGroup;
  submitted = false;
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
        password: ["", [Validators.required, Validators.minLength(8)]],
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
