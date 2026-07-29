/**
 * @typedef {Object} Client
 * @property {string} id
 * @property {string} user_id
 * @property {string} offer_id
 * @property {number} visits
 * @property {number} last_visit
 * @property {number} created_at
 */

/**
 * @typedef {Object} ListClients
 * @property {boolean} has_next_page
 * @property {boolean} has_prev_page
 * @property {number} total_pages
 * @property {Client[]} list
 */

class ClientApi extends Api {
    constructor(clientUrn) {
        super(clientUrn)
    }

    /**
     * @param {string} offerId 
     * @param {Object} visit
     * @param {string} visit.date
     * @param {DayTime} visit.startTime
     * @returns {Promise<Client>}
     */
    async upsert(offerId, visit) {
        return this.request(
            '/client',
            'POST',
            {
                offer_id: offerId,
                visit: {
                    date: visit?.date,
                    start_time: visit?.startTime
                }
            }
        )
    }

    /**
     * @param {string} clientId 
     * @returns {Promise<Client>}
     */
    async getById(clientId) {
        return this.request(
            `/client/${clientId}`,
            'GET'
        )
    }

    /**
     * @param {string} offerId 
     * @param {number} page 
     * @param {number} limit 
     * @returns {Promise<ListClients>}
     */
    async listByOfferId(offerId, page, limit) {
        const query = new URLSearchParams({
            page,
            limit
        })
        return this.request(
            `/offer/${offerId}?${query}`,
            `GET`
        )
    }

    /**
     * @param {number} page 
     * @param {number} limit 
     * @returns {Promise<ListClients>}
     */
    async list(page, limit) {
        const query = new URLSearchParams({
            page,
            limit
        })
        return this.request(
            `/user?${query}`,
            `GET`
        )
    }
}

const clientAPI = new ClientApi('/clients')