import express from 'express';
import { register, login, logout, getCurrentUser, resetPassword } from '../controllers/authController.js';
import { requireAuth } from '../middleware/auth.js';

const router = express.Router();

router.post('/register', register);
router.post('/login', login);
router.post('/logout', logout);
router.get('/me', requireAuth, getCurrentUser);
router.post('/reset-password', resetPassword);

export default router;

