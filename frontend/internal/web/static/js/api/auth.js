/**
 * @typedef {Object} User
 * @property {string} id
 * @property {string} firstname
 * @property {string} lastname
 * @property {string} email
 * @property {string} phone_number
 * @property {number} paid_until
 * @property {boolean} on_vacation
 * @property {number} tg_chat_id
 * @property {Object} notification
 * @property {boolean} notification.telegram_enable
 * @property {boolean} trial_used
 * @property {number} created_at
 */

/**
 * @typedef {Object} TgBinding
 * @property {string} id
 * @property {string} user_id
 * @property {string} bot_url
 * @property {number} created_at
 */

/**
 * @typedef {Object} CodeResent
 * @property {number} next_send_at 
 */

class AuthApi extends Api {
    constructor(authUrn) {
        super(authUrn)
    }

    /**
     * @param {Object} params
     * @param {string} params.firstname
     * @param {string} [params.lastname]
     * @param {string} params.email
     * @param {string} [params.phone_number]
     * @param {string} params.password
     * @returns {Promise<User>}
     */
    async prepare(
        firstname,
        lastname,
        email,
        phone_number,
        password,
    ) {
        return this.request(
            '/prepare',
            'POST',
            {
                firstname: firstname,
                lastname: lastname,
                email: email,
                phone_number: phone_number,
                password: password
            }
        )
    }

    /**
     * @param {string} email 
     * @param {number} code 
     * @returns {Promise<null>}
     */
    async confirm(
        email,
        code,
    ) {
        return this.request(
            '/confirm',
            'POST',
            {
                email: email,
                code: code
            }
        )
    }

    /**
     * @param {string} email 
     * @param {string} password 
     * @return {Promise<null>}
     */
    async login(
        email,
        password
    ) {
        return this.request(
            '/login',
            'POST',
            {
                email: email,
                password: password
            }
        )
    }

    /**
     * @param {string|null} firstname
     * @param {string|null} lastname
     * @param {string|null} phone_number
     * @param {boolean|null} on_vacation
     * @param {object} notification
     * @param {boolean|null} notification.telegram_enable
     * @returns {Promise<null>}
     */
    async update(
        notification,
        firstname,
        lastname,
        phone_number,
        on_vacation,
    ) {
        return this.request(
            '/user',
            'PATCH',
            {
                firstname: firstname,
                lastname: lastname,
                phone_number: phone_number,
                on_vacation: on_vacation,
                notification: {
                    telegram_enable: notification?.telegram_enable
                }
            }
        )
    }

    /**
     * @returns {Promise<User>}
     */
    async self() {
        return this.request(
            '/user',
            'GET'
        )
    }

    /**
     * @returns {Promise<TgBinding>}
     */
    async bind_tg() {
        return this.request(
            '/tgbinding',
            'GET'
        )
    }

    /**
     * @param {string} email 
     * @returns {Promise<CodeResent>}
     */
    async resend_code(email) {
        return this.request(
            '/resend',
            'POST',
            {
                email: email
            }
        )
    }

    /**
     * @param {string} email 
     * @returns {Promise<null>}
     */
    async forget_password(email) {
        return this.request(
            '/forget-password',
            'POST',
            {
                email: email
            }
        )
    }

    /**
     * @param {string} reset_token 
     * @param {string} password 
     * @returns {Promise<null>}
     */
    async change_password(
        reset_token,
        password
    ) {
        return this.request(
            '/password',
            'PATCH',
            {
                reset_token: reset_token,
                new_password: password
            }
        )
    }

    /**
     * @returns {Promise<User>}
     */
    async unbind_tg() {
        return this.request(
            '/unbind-tg',
            'PATCH'
        )
    }
}

const authAPI = new AuthApi('/auth')