import db from '../config/database.js';
import { hashPassword, comparePassword, generateRecoverySecret, hashRecoverySecret, compareRecoverySecret } from '../utils/passwordHash.js';
import { randomUUID } from 'crypto';

export async function register(req, res) {
  try {
    const { username, password } = req.body;

    if (!username || !password) {
      return res.status(400).json({ error: 'Username and password are required' });
    }

    if (password.length < 6) {
      return res.status(400).json({ error: 'Password must be at least 6 characters' });
    }

    // Check if username already exists
    const existingUser = await db.get('SELECT id FROM users WHERE username = ?', username);
    if (existingUser) {
      return res.status(400).json({ error: 'Username already exists' });
    }

    // Generate recovery credentials
    const recoveryUuid = randomUUID();
    const recoverySecret = generateRecoverySecret();
    const recoverySecretHash = await hashRecoverySecret(recoverySecret);

    // Hash password and create user
    const passwordHash = await hashPassword(password);
    const result = await db.run(
      'INSERT INTO users (username, password_hash, recovery_uuid, recovery_secret_hash) VALUES (?, ?, ?, ?)',
      username,
      passwordHash,
      recoveryUuid,
      recoverySecretHash
    );

    // Set session
    req.session.userId = result.lastInsertRowid;
    req.session.username = username;

    res.status(201).json({
      message: 'User registered successfully',
      user: { id: result.lastInsertRowid, username },
      recovery: {
        uuid: recoveryUuid,
        secret: recoverySecret
      }
    });
  } catch (error) {
    console.error('Registration error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function login(req, res) {
  try {
    const { username, password } = req.body;

    if (!username || !password) {
      return res.status(400).json({ error: 'Username and password are required' });
    }

    // Find user
    const user = await db.get('SELECT id, username, password_hash FROM users WHERE username = ?', username);

    if (!user) {
      return res.status(401).json({ error: 'Invalid username or password' });
    }

    // Verify password
    const isValid = await comparePassword(password, user.password_hash);
    if (!isValid) {
      return res.status(401).json({ error: 'Invalid username or password' });
    }

    // Set session
    req.session.userId = user.id;
    req.session.username = user.username;

    res.json({
      message: 'Login successful',
      user: { id: user.id, username: user.username }
    });
  } catch (error) {
    console.error('Login error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function logout(req, res) {
  req.session.destroy((err) => {
    if (err) {
      return res.status(500).json({ error: 'Could not log out' });
    }
    res.clearCookie('connect.sid');
    res.json({ message: 'Logged out successfully' });
  });
}

export async function getCurrentUser(req, res) {
  if (req.session && req.session.userId) {
    const user = await db.get('SELECT id, username, created_at FROM users WHERE id = ?', req.session.userId);
    return res.json({ user });
  }
  return res.status(401).json({ error: 'Not authenticated' });
}

export async function resetPassword(req, res) {
  try {
    const { recoveryUuid, recoverySecret, newPassword } = req.body;

    if (!recoveryUuid || !recoverySecret || !newPassword) {
      return res.status(400).json({ error: 'Recovery UUID, recovery secret, and new password are required' });
    }

    if (newPassword.length < 6) {
      return res.status(400).json({ error: 'Password must be at least 6 characters' });
    }

    // Find user by recovery UUID
    const user = await db.get(
      'SELECT id, username, recovery_uuid, recovery_secret_hash FROM users WHERE recovery_uuid = ?',
      recoveryUuid
    );

    if (!user) {
      return res.status(401).json({ error: 'Invalid recovery credentials' });
    }

    // Verify recovery secret
    const isValid = await compareRecoverySecret(recoverySecret, user.recovery_secret_hash);
    if (!isValid) {
      return res.status(401).json({ error: 'Invalid recovery credentials' });
    }

    // Hash new password and update
    const passwordHash = await hashPassword(newPassword);
    await db.run('UPDATE users SET password_hash = ? WHERE id = ?', passwordHash, user.id);

    res.json({
      message: 'Password reset successfully',
      user: { id: user.id, username: user.username }
    });
  } catch (error) {
    console.error('Password reset error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}
