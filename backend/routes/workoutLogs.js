import express from 'express';
import { requireAuth } from '../middleware/auth.js';
import {
  getAllWorkoutLogs,
  getWorkoutLogById,
  createWorkoutLog,
  updateWorkoutLog,
  deleteWorkoutLog,
  getLastWorkoutValues
} from '../controllers/workoutLogController.js';

const router = express.Router();

// All routes require authentication
router.use(requireAuth);

router.get('/', getAllWorkoutLogs);
router.get('/:id', getWorkoutLogById);
router.post('/', createWorkoutLog);
router.put('/:id', updateWorkoutLog);
router.delete('/:id', deleteWorkoutLog);
router.get('/exercise/:id/last', getLastWorkoutValues);

export default router;

