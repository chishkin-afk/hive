/**
 * @typedef {Object} Offer
 * @property {string} id
 * @property {string} author_id
 * @property {string} title
 * @property {string} description
 * @property {number} price
 * @property {number} slot_duration_minutes
 * @property {number} slot_interval_minutes
 * @property {boolean} is_active
 * @property {number} created_at
 * @property {number} updated_at
 */

/**
 * @typedef {Object} ListOffers
 * @property {boolean} has_next_page
 * @property {boolean} has_prev_page
 * @property {number} total_pages
 * @property {Offer[]} list
 */

class OfferApi extends Api {
    constructor(offerUrn) {
        super(offerUrn)
    }

    /**
     * @param {Object} params
     * @param {string} params.title 
     * @param {string|null} params.description 
     * @param {number} params.pricePenny 
     * @param {number} params.slotDurationMinutes 
     * @param {number} params.slotIntervalMinutes 
     * @returns {Promise<Offer>}
     */
    async create({
        title,
        description,
        pricePenny,
        slotDurationMinutes,
        slotIntervalMinutes,
    }) {
        return this.request(
            '/offer',
            'POST',
            {
                title: title,
                description: description,
                price_penny: pricePenny,
                slot_duration_minutes: slotDurationMinutes,
                slot_interval_minutes: slotIntervalMinutes,
            }
        )
    }

    /**
     * @param {Object} params
     * @param {string} params.offerId
     * @param {string|null} params.title
     * @param {string|null} params.description
     * @param {number|null} params.pricePenny
     * @param {number|null} params.slotDurationMinutes
     * @param {number|null} params.slotIntervalMinutes
     * @param {boolean|null} params.isActive
     * @returns {Promise<Offer>}
     */
    async update({
        offerId,
        title,
        description,
        pricePenny,
        slotDurationMinutes,
        slotIntervalMinutes,
        isActive,
    }) {
        return this.request(
            '/offer',
            'PATCH',
            {
                offer_id: offerId,
                title: title,
                description: description ?? '',
                price_penny: pricePenny,
                slot_duration_minutes: slotDurationMinutes,
                slot_interval_minutes: slotIntervalMinutes,
                is_active: isActive,
            }
        )
    }

    /**
     * @param {string} offerId 
     * @returns {Promise<null>}
     */
    async delete(offerId) {
        return this.request(
            `/offer/${offerId}`,
            'DELETE'
        )
    }

    /**
     * @param {string} offerId 
     * @returns {Promise<Offer>}
     */
    async getById(offerId) {
        return this.request(
            `/offer/${offerId}`,
            'GET'
        )
    }

    /**
     * @param {string} authorId 
     * @param {number} page 
     * @param {number} limit 
     * @returns {Promise<ListOffers>}
     */
    async list(authorId, page, limit) {
        const query = new URLSearchParams({
            page,
            limit,
        });

        return this.request(
            `/user/${authorId}?${query}`,
            'GET'
        )
    }
}

window.offerAPI = new OfferApi('/offers')