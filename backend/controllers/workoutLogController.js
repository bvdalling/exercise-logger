import db from '../config/database.js';

export async function getAllWorkoutLogs(req, res) {
  try {
    const userId = req.session.userId;
    const { exercise_id, limit, start_date, end_date } = req.query;

    let query = `
      SELECT wl.*, e.name as exercise_name, e.exercise_type
      FROM workout_logs wl
      JOIN exercises e ON wl.exercise_id = e.id
      WHERE wl.user_id = ?
    `;
    const params = [userId];

    if (exercise_id) {
      query += ' AND wl.exercise_id = ?';
      params.push(parseInt(exercise_id));
    }

    if (start_date) {
      query += ' AND wl.date >= ?';
      params.push(start_date);
    }

    if (end_date) {
      query += ' AND wl.date <= ?';
      params.push(end_date);
    }

    query += ' ORDER BY wl.date DESC, wl.created_at DESC';

    if (limit) {
      query += ' LIMIT ?';
      params.push(parseInt(limit));
    }

    const logs = await db.all(query, ...params);
    
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
    
    res.json({ logs });
  } catch (error) {
    console.error('Get workout logs error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function getWorkoutLogById(req, res) {
  try {
    const userId = req.session.userId;
    const logId = parseInt(req.params.id);

    const log = await db.get(`
      SELECT wl.*, e.name as exercise_name, e.exercise_type
      FROM workout_logs wl
      JOIN exercises e ON wl.exercise_id = e.id
      WHERE wl.id = ? AND wl.user_id = ?
    `, logId, userId);

    if (!log) {
      return res.status(404).json({ error: 'Workout log not found' });
    }

    // Parse JSON fields
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

    res.json({ log });
  } catch (error) {
    console.error('Get workout log error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function createWorkoutLog(req, res) {
  try {
    const userId = req.session.userId;
    const { exercise_id, date, sets, reps, weight, weight_per_set, rest_time, distance, duration, pace, lap_times, notes } = req.body;

    if (!exercise_id || !date) {
      return res.status(400).json({ error: 'Exercise ID and date are required' });
    }

    // Verify exercise belongs to user and get exercise type
    const exercise = await db.get('SELECT id, exercise_type FROM exercises WHERE id = ? AND user_id = ?', exercise_id, userId);

    if (!exercise) {
      return res.status(404).json({ error: 'Exercise not found' });
    }

    // Validate: distance/time should only be used for cardio
    if (exercise.exercise_type !== 'cardio' && (distance !== undefined && distance !== null || duration !== undefined && duration !== null || pace !== undefined && pace !== null || lap_times !== undefined && lap_times !== null)) {
      return res.status(400).json({ error: 'Distance, duration, pace, and lap times can only be used for cardio exercises' });
    }

    // Validate: weight/weight_per_set should only be used for strength
    if (exercise.exercise_type !== 'strength' && (weight !== undefined && weight !== null || weight_per_set !== undefined && weight_per_set !== null)) {
      return res.status(400).json({ error: 'Weight and weight per set can only be used for strength exercises' });
    }

    // Serialize JSON fields
    const weightPerSetJson = weight_per_set ? JSON.stringify(weight_per_set) : null;
    const lapTimesJson = lap_times ? JSON.stringify(lap_times) : null;

    const result = await db.run(`
      INSERT INTO workout_logs (user_id, exercise_id, date, sets, reps, weight, weight_per_set, rest_time, distance, duration, pace, lap_times, notes)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, userId, exercise_id, date, sets || null, reps || null, weight || null,
        weightPerSetJson, rest_time || null, distance || null, duration || null, pace || null, lapTimesJson, notes || null);

    const log = await db.get(`
      SELECT wl.*, e.name as exercise_name, e.exercise_type
      FROM workout_logs wl
      JOIN exercises e ON wl.exercise_id = e.id
      WHERE wl.id = ?
    `, result.lastInsertRowid);

    // Parse JSON fields
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

    res.status(201).json({ log });
  } catch (error) {
    console.error('Create workout log error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function updateWorkoutLog(req, res) {
  try {
    const userId = req.session.userId;
    const logId = parseInt(req.params.id);
    const { exercise_id, date, sets, reps, weight, weight_per_set, rest_time, distance, duration, pace, lap_times, notes } = req.body;

    // Verify log belongs to user
    const existing = await db.get('SELECT id, exercise_id FROM workout_logs WHERE id = ? AND user_id = ?', logId, userId);

    if (!existing) {
      return res.status(404).json({ error: 'Workout log not found' });
    }

    // Get exercise type for validation
    const currentExerciseId = exercise_id !== undefined ? exercise_id : existing.exercise_id;
    const exercise = await db.get('SELECT id, exercise_type FROM exercises WHERE id = ? AND user_id = ?', currentExerciseId, userId);
    
    if (!exercise) {
      return res.status(404).json({ error: 'Exercise not found' });
    }

    // Validate: distance/time should only be used for cardio
    if (exercise.exercise_type !== 'cardio' && (distance !== undefined && distance !== null || duration !== undefined && duration !== null || pace !== undefined && pace !== null || lap_times !== undefined && lap_times !== null)) {
      return res.status(400).json({ error: 'Distance, duration, pace, and lap times can only be used for cardio exercises' });
    }

    // Validate: weight/weight_per_set should only be used for strength
    if (exercise.exercise_type !== 'strength' && (weight !== undefined && weight !== null || weight_per_set !== undefined && weight_per_set !== null)) {
      return res.status(400).json({ error: 'Weight and weight per set can only be used for strength exercises' });
    }

    // Build update query dynamically based on provided fields
    const updates = []
    const values = []
    
    if (exercise_id !== undefined) {
      updates.push('exercise_id = ?')
      values.push(exercise_id)
    }
    if (date !== undefined) {
      updates.push('date = ?')
      values.push(date)
    }
    if (sets !== undefined) {
      updates.push('sets = ?')
      values.push(sets || null)
    }
    if (reps !== undefined) {
      updates.push('reps = ?')
      values.push(reps || null)
    }
    if (weight !== undefined) {
      updates.push('weight = ?')
      values.push(weight || null)
    }
    if (weight_per_set !== undefined) {
      updates.push('weight_per_set = ?')
      values.push(weight_per_set ? JSON.stringify(weight_per_set) : null)
    }
    if (rest_time !== undefined) {
      updates.push('rest_time = ?')
      values.push(rest_time || null)
    }
    if (distance !== undefined) {
      updates.push('distance = ?')
      values.push(distance || null)
    }
    if (duration !== undefined) {
      updates.push('duration = ?')
      values.push(duration || null)
    }
    if (pace !== undefined) {
      updates.push('pace = ?')
      values.push(pace || null)
    }
    if (lap_times !== undefined) {
      updates.push('lap_times = ?')
      values.push(lap_times ? JSON.stringify(lap_times) : null)
    }
    if (notes !== undefined) {
      updates.push('notes = ?')
      values.push(notes || null)
    }
    
    if (updates.length > 0) {
      values.push(logId, userId)
      await db.run(`
        UPDATE workout_logs
        SET ${updates.join(', ')}
        WHERE id = ? AND user_id = ?
      `, ...values)
    }

    const log = await db.get(`
      SELECT wl.*, e.name as exercise_name, e.exercise_type
      FROM workout_logs wl
      JOIN exercises e ON wl.exercise_id = e.id
      WHERE wl.id = ?
    `, logId);

    // Parse JSON fields
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

    res.json({ log });
  } catch (error) {
    console.error('Update workout log error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function deleteWorkoutLog(req, res) {
  try {
    const userId = req.session.userId;
    const logId = parseInt(req.params.id);

    // Verify log belongs to user
    const existing = await db.get('SELECT id FROM workout_logs WHERE id = ? AND user_id = ?', logId, userId);

    if (!existing) {
      return res.status(404).json({ error: 'Workout log not found' });
    }

    await db.run('DELETE FROM workout_logs WHERE id = ? AND user_id = ?', logId, userId);
    res.json({ message: 'Workout log deleted successfully' });
  } catch (error) {
    console.error('Delete workout log error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}

export async function getLastWorkoutValues(req, res) {
  try {
    const userId = req.session.userId;
    const exerciseId = parseInt(req.params.id);

    // Verify exercise belongs to user
    const exercise = await db.get('SELECT id FROM exercises WHERE id = ? AND user_id = ?', exerciseId, userId);

    if (!exercise) {
      return res.status(404).json({ error: 'Exercise not found' });
    }

    // Get the most recent workout log for this exercise
    const lastLog = await db.get(`
      SELECT sets, reps, weight, weight_per_set, rest_time, distance, duration, pace, lap_times, date
      FROM workout_logs
      WHERE exercise_id = ? AND user_id = ?
      ORDER BY date DESC, created_at DESC
      LIMIT 1
    `, exerciseId, userId);

    // Parse JSON fields
    if (lastLog) {
      if (lastLog.weight_per_set) {
        try {
          lastLog.weight_per_set = JSON.parse(lastLog.weight_per_set);
        } catch (e) {
          lastLog.weight_per_set = null;
        }
      }
      if (lastLog.lap_times) {
        try {
          lastLog.lap_times = JSON.parse(lastLog.lap_times);
        } catch (e) {
          lastLog.lap_times = null;
        }
      }
    }

    res.json({ lastLog: lastLog || null });
  } catch (error) {
    console.error('Get last workout values error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}
