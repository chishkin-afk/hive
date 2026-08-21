/**
 * @typedef {Object} Booking
 * @property {string} id
 * @property {string} offer_id
 * @property {string} user_id
 * @property {string} comment
 * @property {number} start_time
 * @property {boolean} is_cancelled
 * @property {number} created_at
 * @property {number} updated_at
 */

/**
 * @typedef {Object} BookingUser
 * @property {string} id
 * @property {string} firstname
 * @property {string} lastname
 * @property {string} confirmed_email
 */

/**
 * @typedef {Object} ListBookingElement
 * @property {Booking} booking
 * @property {BookingUser} user
 */

/**
 * @typedef {Object} ListBookings
 * @property {ListBookingElement[]} list
 * @property {number} total_pages
 * @property {boolean} has_next_page
 * @property {boolean} has_prev_page
 */

/**
 * @typedef {Object} Slots
 * @property {string} date
 * @property {DayTime[]} times
 */

/**
 * @typedef {Object} Day
 * @property {number} time
 * @property {boolean} is_active
 */

/**
 * @typedef {Object} ListDays
 * @property {Day[]} list
 */

class BookingApi extends Api {
    constructor(offerUrn) {
        super(offerUrn)
    }

    /**
     * 
     * @param {Object} params
     * @param {string} params.offerId
     * @param {string} params.comment
     * @param {Object} params.startTime
     * @param {string} params.startTime.date
     * @param {DayTime} params.startTime.start_time
     * @returns 
     */
    async create({
        offerId,
        comment,
        startTime
    }) {   
        return this.request(
            '/booking',
            'POST',
            {
                offer_id: offerId,
                comment: comment,
                start_time: startTime
            }
        )
    }

    /**
     * @param {string} bookingId 
     * @returns {Promise<Booking>}
     */
    async getById(bookingId) {
        return this.request(
            `/booking/${bookingId}`,
            'GET'
        )
    }

    /**
     * @param {number} page 
     * @param {number} limit 
     * @param {string|null} offerId
     * @returns {Promise<ListBookings>}
     */
    async list(page, limit, offerId) {
        const query = new URLSearchParams()
        query.set("page", page)
        query.set("limit", limit)

        if (offerId) {
            query.set("offer_id", offerId)
        }

        return this.request(
            `/bookings?${query}`,
            'GET'
        )
    }

    /**
     * @param {Object} params
     * @param {string} params.bookingId
     * @param {boolean} params.isCancelled
     * @param {string} params.comment
     * @param {Object|null} params.startTime
     * @param {string} params.startTime.date
     * @param {DayTime} params.startTime.start_time
     * @returns {Promise<Booking>}
     */
    async update({
        bookingId,
        isCancelled,
        comment,
        startTime
    }) {
        return this.request(
            '/booking',
            'PATCH',
            {
                booking_id: bookingId,
                is_cancelled: isCancelled,
                comment: comment,
                start_time: startTime
            }
        )
    }

    /**
     * @param {string} offerId 
     * @param {string} date  // YYYY-MM-DD
     * @returns {Promise<Slots>}
     */
    async slots(offerId, date) {
        const query = new URLSearchParams({
            date
        })

        return this.request(
            `/slots/${offerId}?${query}`,
            'GET'
        )
    }

    /**
     * @param {string} offerId 
     * @param {string} from 
     * @param {string} to 
     * @returns {Promise<ListDays>}
     */
    async allDays(offerId, from, to) {
        const query = new URLSearchParams({
            from,
            to
        })

        return this.request(
            `/days/${offerId}?${query}`,
            'GET'
        )
    }
}

window.bookingAPI = new BookingApi('/bookings')