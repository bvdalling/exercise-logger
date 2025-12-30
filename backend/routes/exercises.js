import express from 'express';
import { requireAuth } from '../middleware/auth.js';
import {
  getAllExercises,
  getExerciseById,
  createExercise,
  updateExercise,
  deleteExercise,
  getExerciseProgress
} from '../controllers/exerciseController.js';

const router = express.Router();

// All routes require authentication
router.use(requireAuth);

router.get('/', getAllExercises);
router.get('/:id', getExerciseById);
router.post('/', createExercise);
router.put('/:id', updateExercise);
router.delete('/:id', deleteExercise);
router.get('/:id/progress', getExerciseProgress);

export default router;

