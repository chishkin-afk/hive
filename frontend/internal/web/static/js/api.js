class HttpClient {
  constructor() {
    this.baseUrl = `${window.APP_CONFIG?.apiUrl || ''}`;
  }

  async request(method, path, body) {
    const token = localStorage.getItem('auth_jwt_token');
    const headers = { 'Content-Type': 'application/json' };
    
    if (token) headers['Authorization'] = token;

    const response = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (response.status === 401) {
      localStorage.removeItem('auth_jwt_token');
      window.location.href = '/login';
      throw new Error('Unauthorized');
    }

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `HTTP Error ${response.status}`);
    }

    return response.status === 204 ? null : response.json();
  }

  get(path) { return this.request('GET', path); }
  post(path, body) { return this.request('POST', path, body); }
  patch(path, body) { return this.request('PATCH', path, body); }
}

class AuthService {
  constructor() {
    this.http = new HttpClient();
  }

  _saveToken(token) {
    localStorage.setItem('auth_jwt_token', token);
  }

  logout() {
    localStorage.removeItem('auth_jwt_token');
  }

  prepareUser(data) {
    return this.http.post('/auth/prepare', data);
  }

  async confirmUser(data) {
    const response = await this.http.post('/auth/confirm', data);
    this._saveToken(response.token);
    return response;
  }

  async login(data) {
    const response = await this.http.post('/auth/login', data);
    this._saveToken(response.token);
    return response;
  }

  getCurrentUser() {
    return this.http.get('/auth/user');
  }

  updateCurrentUser(data) {
    return this.http.patch('/auth/user', data);
  }

  getTelegramBinding() {
    return this.http.get('/auth/tgbinding');
  }
}

class PaymentsService {
  constructor() {
    this.http = new HttpClient();
  }

  getRates() {
    return this.http.get('/payments/rates');
  }

  getRate(id) {
    return this.http.get(`/payments/rate/${id}`);
  }

  createPayment(rateId) {
    return this.http.post('/payments/payment', { rate_id: rateId });
  }

  getPayment(id) {
    return this.http.get(`/payments/payment/${id}`);
  }

  getUserPayments({ page = 1, limit = 100 } = {}) {
    const query = new URLSearchParams({ page, limit }).toString();
    return this.http.get(`/payments/user?${query}`);
  }

  takeTrial() {
    return this.http.post('/payments/take-trial');
  }
}

const authService = new AuthService();
const paymentsService = new PaymentsService();