import db from '../config/database.js';

export async function getAllExercises(req, res) {
  try {
    const userId = req.session.userId;
    const exercises = await db.all(`
      SELECT * FROM exercises 
      WHERE user_id = ? 
      ORDER BY created_at DESC
    `, userId);

    res.json({ exercises });
  } catch (error) {
    console.error('Get exercises error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function getExerciseById(req, res) {
  try {
    const userId = req.session.userId;
    const exerciseId = parseInt(req.params.id);

    const exercise = await db.get(`
      SELECT * FROM exercises 
      WHERE id = ? AND user_id = ?
    `, exerciseId, userId);

    if (!exercise) {
      return res.status(404).json({ error: 'Exercise not found' });
    }

    res.json({ exercise });
  } catch (error) {
    console.error('Get exercise error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function createExercise(req, res) {
  try {
    const userId = req.session.userId;
    const { name, exercise_type, muscle_group, equipment, description, instructions, video_link, image_link } = req.body;

    if (!name) {
      return res.status(400).json({ error: 'Exercise name is required' });
    }

    // Validate exercise_type
    const validTypes = ['strength', 'cardio'];
    const type = exercise_type && validTypes.includes(exercise_type) ? exercise_type : 'strength';

    const result = await db.run(`
      INSERT INTO exercises (user_id, name, exercise_type, muscle_group, equipment, description, instructions, video_link, image_link)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, userId, name, type, muscle_group || null, equipment || null, description || null,
        instructions || null, video_link || null, image_link || null);

    const exercise = await db.get('SELECT * FROM exercises WHERE id = ?', result.lastInsertRowid);

    res.status(201).json({ exercise });
  } catch (error) {
    console.error('Create exercise error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function updateExercise(req, res) {
  try {
    const userId = req.session.userId;
    const exerciseId = parseInt(req.params.id);
    const { name, exercise_type, muscle_group, equipment, description, instructions, video_link, image_link } = req.body;

    // Verify exercise belongs to user
    const existing = await db.get('SELECT id FROM exercises WHERE id = ? AND user_id = ?', exerciseId, userId);

    if (!existing) {
      return res.status(404).json({ error: 'Exercise not found' });
    }

    // Validate exercise_type if provided
    let type = null;
    if (exercise_type !== undefined) {
      const validTypes = ['strength', 'cardio'];
      type = validTypes.includes(exercise_type) ? exercise_type : 'strength';
    }

    // Build update query dynamically
    const updates = [];
    const values = [];
    
    if (name !== undefined) {
      updates.push('name = ?');
      values.push(name);
    }
    if (type !== null) {
      updates.push('exercise_type = ?');
      values.push(type);
    }
    if (muscle_group !== undefined) {
      updates.push('muscle_group = ?');
      values.push(muscle_group || null);
    }
    if (equipment !== undefined) {
      updates.push('equipment = ?');
      values.push(equipment || null);
    }
    if (description !== undefined) {
      updates.push('description = ?');
      values.push(description || null);
    }
    if (instructions !== undefined) {
      updates.push('instructions = ?');
      values.push(instructions || null);
    }
    if (video_link !== undefined) {
      updates.push('video_link = ?');
      values.push(video_link || null);
    }
    if (image_link !== undefined) {
      updates.push('image_link = ?');
      values.push(image_link || null);
    }

    if (updates.length > 0) {
      values.push(exerciseId, userId);
      await db.run(`
        UPDATE exercises 
        SET ${updates.join(', ')}
        WHERE id = ? AND user_id = ?
      `, ...values);
    }

    const exercise = await db.get('SELECT * FROM exercises WHERE id = ?', exerciseId);
    res.json({ exercise });
  } catch (error) {
    console.error('Update exercise error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function deleteExercise(req, res) {
  try {
    const userId = req.session.userId;
    const exerciseId = parseInt(req.params.id);

    // Verify exercise belongs to user
    const existing = await db.get('SELECT id FROM exercises WHERE id = ? AND user_id = ?', exerciseId, userId);

    if (!existing) {
      return res.status(404).json({ error: 'Exercise not found' });
    }

    await db.run('DELETE FROM exercises WHERE id = ? AND user_id = ?', exerciseId, userId);
    res.json({ message: 'Exercise deleted successfully' });
  } catch (error) {
    console.error('Delete exercise error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function getExerciseProgress(req, res) {
  try {
    const userId = req.session.userId;
    const exerciseId = parseInt(req.params.id);

    // Verify exercise belongs to user
    const exercise = await db.get('SELECT id FROM exercises WHERE id = ? AND user_id = ?', exerciseId, userId);

    if (!exercise) {
      return res.status(404).json({ error: 'Exercise not found' });
    }

    // Get all workout logs for this exercise, ordered by date
    const logs = await db.all(`
      SELECT id, date, weight, weight_per_set, rest_time, distance, duration, pace, lap_times, sets, reps, notes
      FROM workout_logs
      WHERE exercise_id = ? AND user_id = ?
      ORDER BY date ASC
    `, exerciseId, userId);

    // Parse JSON fields for all logs
    logs.forEach(log => {
      if (log.weight_per_set) {
        try {
          log.weight_per_set = JSON.parse(log.weight_per_set);
        } catch (e) {
          log.weight_per_set = null;
        }
      }
      if (log.lap_times) {
        try {
          log.lap_times = JSON.parse(log.lap_times);
        } catch (e) {
          log.lap_times = null;
        }
      }
    });

    res.json({ progress: logs });
  } catch (error) {
    console.error('Get progress error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}
