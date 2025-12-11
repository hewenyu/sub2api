export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

export function isValidUsername(username: string): boolean {
  if (!username || username.length < 3 || username.length > 32) {
    return false;
  }

  const usernameRegex = /^[a-zA-Z0-9_-]+$/;
  return usernameRegex.test(username);
}

export function isValidPassword(password: string): boolean {
  return password.length >= 6;
}

export function isValidApiKey(apiKey: string): boolean {
  return apiKey.length >= 10;
}

export interface ValidationError {
  field: string;
  message: string;
}

export function validateLoginForm(username: string, password: string): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!username) {
    errors.push({ field: 'username', message: 'Username is required' });
  } else if (!isValidUsername(username)) {
    errors.push({
      field: 'username',
      message:
        'Username must be 3-32 characters and contain only letters, numbers, hyphens, and underscores',
    });
  }

  if (!password) {
    errors.push({ field: 'password', message: 'Password is required' });
  } else if (!isValidPassword(password)) {
    errors.push({
      field: 'password',
      message: 'Password must be at least 6 characters',
    });
  }

  return errors;
}

export function validateAPIKeyForm(
  keyName: string,
  apiKey: string,
  provider: string
): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!keyName || keyName.trim().length === 0) {
    errors.push({ field: 'keyName', message: 'Key name is required' });
  }

  if (!apiKey || apiKey.trim().length === 0) {
    errors.push({ field: 'apiKey', message: 'API key is required' });
  } else if (!isValidApiKey(apiKey)) {
    errors.push({
      field: 'apiKey',
      message: 'API key must be at least 10 characters',
    });
  }

  if (!provider || provider.trim().length === 0) {
    errors.push({ field: 'provider', message: 'Provider is required' });
  }

  return errors;
}

export function validateCodexAccountForm(email: string, password: string): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!email) {
    errors.push({ field: 'email', message: 'Email is required' });
  } else if (!isValidEmail(email)) {
    errors.push({ field: 'email', message: 'Invalid email format' });
  }

  if (!password) {
    errors.push({ field: 'password', message: 'Password is required' });
  } else if (!isValidPassword(password)) {
    errors.push({
      field: 'password',
      message: 'Password must be at least 6 characters',
    });
  }

  return errors;
}

export function validateAdminForm(
  username: string,
  email: string,
  password: string
): ValidationError[] {
  const errors: ValidationError[] = [];

  if (!username) {
    errors.push({ field: 'username', message: 'Username is required' });
  } else if (!isValidUsername(username)) {
    errors.push({
      field: 'username',
      message:
        'Username must be 3-32 characters and contain only letters, numbers, hyphens, and underscores',
    });
  }

  if (!email) {
    errors.push({ field: 'email', message: 'Email is required' });
  } else if (!isValidEmail(email)) {
    errors.push({ field: 'email', message: 'Invalid email format' });
  }

  if (!password) {
    errors.push({ field: 'password', message: 'Password is required' });
  } else if (!isValidPassword(password)) {
    errors.push({
      field: 'password',
      message: 'Password must be at least 6 characters',
    });
  }

  return errors;
}
