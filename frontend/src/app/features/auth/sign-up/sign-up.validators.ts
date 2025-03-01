import { FormGroup } from '@angular/forms';

export function passwordMatchValidator(form: FormGroup) {
  const password = form.get('password');
  const confirmPassword = form.get('confirmPassword');

  if (password && confirmPassword && password.value !== confirmPassword.value) {
    confirmPassword.setErrors({ passwordMismatch: true });
  } else {
    confirmPassword?.setErrors(null);
  }
}
