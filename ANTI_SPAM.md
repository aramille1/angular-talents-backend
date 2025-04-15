# Anti-Spam Protection for Angular Talents

This document describes the anti-spam measures implemented in the Angular Talents platform to protect against automated bot signups.

## Implemented Protections

### 1. reCAPTCHA v3

The system uses Google's reCAPTCHA v3, which provides invisible protection without requiring user interaction. It assigns a score to each interaction based on how likely it is that the interaction is coming from a legitimate human user.

#### How it works:

1. The frontend sends a token with each signup request
2. The backend verifies this token with Google's API
3. Google returns a score between 0.0 and 1.0 (1.0 being most likely human)
4. Our system rejects submissions with scores below 0.5 (configurable threshold)

#### Setup Instructions:

1. Register for reCAPTCHA v3 at https://www.google.com/recaptcha/admin
2. Add your domain to the allowed domains list
3. Get your Site Key and Secret Key
4. Add these keys to your .env file:
   ```
   RECAPTCHA_SITE_KEY=your_site_key
   RECAPTCHA_SECRET_KEY=your_secret_key
   ```

5. Add the reCAPTCHA script to your Angular application's index.html:
   ```html
   <script src="https://www.google.com/recaptcha/api.js?render=YOUR_SITE_KEY"></script>
   ```

6. Implement in your signup component:
   ```typescript
   // Inside your signup component
   submitForm() {
     grecaptcha.ready(() => {
       grecaptcha.execute('YOUR_SITE_KEY', {action: 'signup'})
         .then(token => {
           // Add token to your form data
           this.signupForm.value.recaptchaToken = token;
           // Submit the form
           this.authService.signup(this.signupForm.value).subscribe(...);
         });
     });
   }
   ```

### 2. Honeypot Field

A hidden field is added to the signup form that should remain empty. Bots often fill all form fields, allowing us to identify and block them.

#### How it works:

1. The form includes a field named "interests" that is hidden with CSS
2. Human users cannot see or fill this field
3. Bots typically fill all fields including this hidden one
4. The backend checks if this field is empty - if not, it silently rejects the submission

#### Implementation:

In your Angular component HTML:
```html
<div class="honeypot-field">
  <input type="text" name="interests" formControlName="interests">
</div>
```

In your CSS:
```css
.honeypot-field {
  display: none;
}
```

In your component:
```typescript
this.signupForm = this.formBuilder.group({
  email: ['', [Validators.required, Validators.email]],
  password: ['', [Validators.required, Validators.minLength(8)]],
  interests: [''] // Honeypot field
});
```

## Best Practices

1. **Don't reveal rejection reasons** - When rejecting bot submissions, don't disclose the exact reason they were identified as bots
2. **Monitor effectiveness** - Regularly check logs for bot attempts and adjust thresholds if needed
3. **Keep reCAPTCHA keys secure** - Never expose your reCAPTCHA secret key in client-side code
4. **Combine with rate limiting** - Consider implementing API rate limiting as an additional layer of protection

## Troubleshooting

If legitimate users are being blocked:

1. Check the reCAPTCHA score threshold - it may be set too high
2. Ensure the honeypot field is properly hidden via CSS
3. Review error logs to identify patterns in false positives
4. Consider temporarily disabling one protection method to isolate the issue
