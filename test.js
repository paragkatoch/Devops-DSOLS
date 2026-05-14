import http from 'k6/http';
import { check, sleep } from 'k6';

const DURATION = __ENV.K6_DURATION || '5m';
const VUS = parseInt(__ENV.K6_VUS || '10');

export const options = {
    scenarios: {
        normal_traffic: {
            executor: 'constant-vus',
            vus: VUS,
            duration: DURATION,
            exec: 'normalTraffic',
        },
        distributed_traffic: {
            executor: 'constant-vus',
            vus: VUS,
            duration: DURATION,
            exec: 'distributedTraffic',
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.10'],
        http_req_duration: ['p(95)<10000'],
    },
};


export default function () {
    normalTraffic();
}


const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1';

// ----------------------------------
// HELPERS
// ----------------------------------

function randomInt(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

function randomChoice(arr) {
    return arr[Math.floor(Math.random() * arr.length)];
}

function randomEmail() {
    return `user${randomInt(1000, 999999)}@gmail.com`;
}

// ----------------------------------
// STATIC IDS
// ----------------------------------

const productIds = ['1', '2'];
const userIds = ['1', '2'];

// ----------------------------------
// USER API
// ----------------------------------

function createUser() {
    const payload = JSON.stringify({
        id: `${randomInt(100, 999999)}`,
        email: randomEmail(),
    });

    return http.post(
        `${BASE_URL}/api/user`,
        payload,
        {
            headers: {
                'Content-Type': 'application/json',
            },
            tags: {
                name: 'Create User',
            },
        }
    );
}

function getUsers() {
    return http.get(
        `${BASE_URL}/api/user`,
        {
            tags: {
                name: 'Get Users',
            },
        }
    );
}

// ----------------------------------
// PRODUCT API
// ----------------------------------

function getProducts() {
    return http.get(
        `${BASE_URL}/api/product`,
        {
            tags: {
                name: 'Get Products',
            },
        }
    );
}

function getProductById(id) {
    return http.get(
        `${BASE_URL}/api/product/${id}`,
        {
            tags: {
                name: 'Get Product By Id',
            },
        }
    );
}

function updateProductQuantity() {

    const productId = randomChoice(productIds);

    // Random inventory fluctuation
    // Sometimes stock added
    // Sometimes removed

    let quantityChange;

    // 20% chance of heavy deduction
    if (Math.random() < 0.2) {
        quantityChange = randomInt(-200, -50);
    }

    // 30% chance of heavy restock
    else if (Math.random() < 0.5) {
        quantityChange = randomInt(50, 300);
    }

    // Normal fluctuation
    else {
        quantityChange = randomInt(-20, 20);
    }

    const payload = JSON.stringify({
        id: productId,
        quantity: quantityChange,
    });

    return http.post(
        `${BASE_URL}/api/product/quantity`,
        payload,
        {
            headers: {
                'Content-Type': 'application/json',
            },
            tags: {
                name: 'Update Product Quantity',
            },
        }
    );
}

// ----------------------------------
// ORDER API
// ----------------------------------

function createOrder() {

    const userId = randomChoice(userIds);

    // Realistic traffic:
    // most orders are small
    // some are huge and fail intentionally

    const quantity1 =
        Math.random() < 0.15
            ? randomInt(100, 400)
            : randomInt(1, 30);

    const quantity2 =
        Math.random() < 0.15
            ? randomInt(100, 400)
            : randomInt(1, 30);

    const payload = JSON.stringify({
        user_id: userId,
        total_amount: 1000,
        currency: 'INR',
        items: [
            {
                product_id: '1',
                quantity: quantity1,
            },
            {
                product_id: '2',
                quantity: quantity2,
            },
        ],
    });

    return http.post(
        `${BASE_URL}/api/order`,
        payload,
        {
            headers: {
                'Content-Type': 'application/json',
            },
            tags: {
                name: 'Create Order',
            },
        }
    );
}

function getOrders() {
    return http.get(
        `${BASE_URL}/api/order`,
        {
            tags: {
                name: 'Get Orders',
            },
        }
    );
}

// ----------------------------------
// LOAD SCENARIOS
// ----------------------------------

export function normalTraffic() {
    const action = randomInt(1, 100);
    let res;

    // GET USERS (25%)
    if (action <= 25) {
        res = getUsers();
        check(res, { 'get users status 200': (r) => r.status === 200 });
    }
    // GET PRODUCTS (50%)
    else if (action <= 75) {
        res = getProducts();
        check(res, { 'get products status 200': (r) => r.status === 200 });
    }
    // GET SINGLE PRODUCT (25%)
    else {
        const productId = randomChoice(productIds);
        res = getProductById(productId);
        check(res, { 'get single product status 200': (r) => r.status === 200 });
    }

    sleep(randomInt(1, 3));
}

export function distributedTraffic() {
    const action = randomInt(1, 100);
    let res;

    // CREATE USER (10%)
    if (action <= 10) {
        res = createUser();
        check(res, { 'create user status 200': (r) => r.status === 200 });
    }
    // UPDATE INVENTORY (40%)
    else if (action <= 50) {
        res = updateProductQuantity();
        check(res, { 'inventory update accepted': (r) => r.status === 200 });
    }
    // CREATE ORDER (40%)
    else if (action <= 90) {
        res = createOrder();
        check(res, { 'order request accepted': (r) => r.status === 200 });
    }
    // GET ORDERS (10%)
    else {
        res = getOrders();
        check(res, { 'get orders status 200': (r) => r.status === 200 });
    }

    sleep(randomInt(1, 3));
}