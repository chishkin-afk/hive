/**
 * @typedef {Object} Payment
 * @property {string} id
 * @property {string} payment_id
 * @property {string} user_id
 * @property {number} amount_penny
 * @property {string} status_string
 * @property {number} status_int
 * @property {number} created_at
 * @property {number} updated_at
 */

/**
 * @typedef {Object} UnconfirmedPayment
 * @property {Payment} payment
 * @property {string} url
 */

/**
 * @typedef {Object} ListPayments
 * @property {Payment[]} list
 * @property {boolean} has_next_page
 * @property {boolean} has_prev_page
 * @property {number} total_pages
 */

/**
 * @typedef {Object} Rate
 * @property {string} id
 * @property {string} name
 * @property {number} amount_penny
 * @property {number} paid_days
 */

/**
 * @typedef {Object} ListRates
 * @property {Rate[]} list
 */

class PaymentApi extends Api {
    constructor(paymentUrn) {
        super(paymentUrn)
    }

    /**
     * @param {string} rateId 
     * @returns {Promise<UnconfirmedPayment>}
     */
    async create(rateId) {
        return this.request(
            '/payment',
            'POST',
            {
                rate_id: rateId
            }
        )
    }

    /**
     * @param {number} page 
     * @param {number} limit 
     * @returns {Promise<ListPayments>}
     */
    async listPayments(page, limit) {
        const query = new URLSearchParams({
            page,
            limit
        })

        return this.request(
            `/user?${query}`,
            'GET'
        )
    }

    /**
     * @param {string} paymentId 
     * @returns {Promise<Payment>}
     */
    async getPaymentById(paymentId) {
        return this.request(
            `/payment/${paymentId}`,
            'GET'
        )
    }

    /**
     * @returns {Promise<ListRates>}
     */
    async listRates() {
        return this.request(
            '/rates',
            'GET'
        )
    }

    /**
     * @param {string} rateId 
     * @returns {Promise<Rate>}
     */
    async getRateById(rateId) {
        return this.request(
            `/rate/${rateId}`,
            'GET'
        )
    }

    /**
     * @returns {Promise<null>}
     */
    async takeTrial() {
        return this.request(
            '/take-trial',
            'PATCH'
        )
    }
}

window.paymentAPI = new PaymentApi('/payments')