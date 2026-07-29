/**
 * @typedef {Object} Schedule
 * @property {string} id
 * @property {string} user_id
 * @property {number} day_of_week
 * @property {number} day_of_week
 * @property {DayTime} start_time
 * @property {DayTime} end_time
 * @property {Break[]} breaks
 */

/**
 * @typedef {Object} Break
 * @property {DayTime} start_time
 * @property {DayTime} end_time
 */

/**
 * @typedef {Object} ListSchedules
 * @property {Schedule[]} list
 */

class ScheduleApi extends Api {
    constructor(offerUrn) {
        super(offerUrn)
    }

    /**
     * @param {Object} params
     * @param {number} params.dayOfWeek
     * @param {DayTime} params.startTime
     * @param {DayTime} params.endTime
     * @param {Break[]} params.breaks
     * @returns {Promise<Schedule>}
     */
    async create({
        dayOfWeek,
        startTime,
        endTime,
        breaks
    }) {
        return this.request(
            '/schedule',
            'POST',
            {
                day_of_week: dayOfWeek,
                start_time: startTime,
                end_time: endTime,
                breaks: breaks
            }
        )
    }

    /**
     * @param {Object} params
     * @param {string} params.scheduleId
     * @param {DayTime|null} params.startTime
     * @param {DayTime|null} params.endTime
     * @param {boolean} params.doUpdateBreak
     * @param {Break[]|null} params.breaks
     * @returns {Promise<Schedule>}
     */
    async update({
        scheduleId,
        startTime,
        endTime,
        doUpdateBreak,
        breaks
    }) {
        return this.request(
            '/schedule',
            'PATCH',
            {
                schedule_id: scheduleId,
                start_time: startTime,
                endTime: endTime,
                do_update_break: doUpdateBreak,
                breaks: breaks
            }
        )
    }

    /**
     * @param {string} userId 
     * @param {number} dayOfWeek 
     * @returns {Promise<Schedule>}
     */
    async getByUserId(userId, dayOfWeek) {
        return this.request(
            `/schedule/${userId}/d/${dayOfWeek}`,
            'GET'
        )
    }

    /**
     * @param {string} userId 
     * @returns {Promise<ListSchedules>}
     */
    async listByUserId(userId) {
        return this.request(
            `/user/${userId}`,
            'GET'
        )
    }

    /**
     * @param {string} scheduleId 
     * @returns {Promise<null>}
     */
    async delete(scheduleId) {
        return this.request(
            `/schedule/${scheduleId}`,
            'DELETE'
        )
    }
}

const scheduleAPI = new ScheduleApi('/schedules')