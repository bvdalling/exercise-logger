import bcrypt from 'bcrypt';
import crypto from 'crypto';

const SALT_ROUNDS = 10;

export async function hashPassword(password) {
  return await bcrypt.hash(password, SALT_ROUNDS);
}

export async function comparePassword(password, hash) {
  return await bcrypt.compare(password, hash);
}

/**
 * Generates a cryptographically secure 32-character recovery secret
 * Uses alphanumeric characters (A-Z, a-z, 0-9)
 */
export function generateRecoverySecret() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  const randomBytes = crypto.randomBytes(32);
  let secret = '';
  
  for (let i = 0; i < 32; i++) {
    secret += chars[randomBytes[i] % chars.length];
  }
  
  return secret;
}

/**
 * Hash the recovery secret using bcrypt
 */
export async function hashRecoverySecret(secret) {
  return await bcrypt.hash(secret, SALT_ROUNDS);
}

/**
 * Compare recovery secret with hash
 */
export async function compareRecoverySecret(secret, hash) {
  return await bcrypt.compare(secret, hash);
}

