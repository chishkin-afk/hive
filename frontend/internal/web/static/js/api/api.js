// Api is a main class of api's
class Api {
    constructor(extendsUrn) {
        console.log(window.APP_CONFIG.apiUrl)
        this._apiURL = window.APP_CONFIG.apiUrl + (extendsUrn ?? '')
    }

    /**
     * @param {string} urn
     * @param {"GET"|"POST"|"PUT"|"PATCH"|"DELETE"} method
     * @param {object=} body
     * @returns {Promise<any>}
     */
    async request(urn, method = "GET", body) {
        const resp = await fetch(this._apiURL + urn, {
            headers: {
                "Content-Type": "application/json"
            },
            credentials: "include",
            method: method,
            body: body !== undefined 
                ? JSON.stringify(body)
                : null
        })

        const text = await resp.text();
        const data = text ? JSON.parse(text) : null;

        if (!resp.ok) {
            const err = new Error(data?.error ?? "unknown error")
            err.status = resp.status
            throw err
        }


        return data
    }
}

