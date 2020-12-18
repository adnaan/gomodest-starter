import { Controller } from "stimulus"

export default class extends Controller {
    static values = { price: String, stripe: String }

    connect(){
        this.stripe = Stripe(this.stripeValue);
    }

    showErrorMessage(message) {
        let errorEl = document.getElementById("error-message")
        errorEl.textContent = message;
        errorEl.style.display = "block";
    };

    handleFetchResult(result) {
        const self = this;
        if (!result.ok) {
            return result.json().then(function(json) {
                if (json.error && json.error.message) {
                    throw new Error(result.url + ' ' + result.status + ' ' + json.error.message);
                }
            }).catch(function(err) {
                self.showErrorMessage(err);
                throw err;
            });
        }
        return result.json();
    };

    handleResult(result) {
        if (result.error) {
            this.showErrorMessage(result.error.message);
        }
    };

    createCheckoutSession(priceId) {
        const self = this;
        return fetch("/account/checkout", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                price: priceId
            })
        }).then(self.handleFetchResult);
    };


    createCheckoutClick(){
        const self = this;
        this.createCheckoutSession(this.priceValue).then(function(data) {
            // Call Stripe.js method to redirect to the new Checkout page
            self.stripe
                .redirectToCheckout({
                    sessionId: data.sessionId
                })
                .then(self.handleResult);
        });
        //console.log("values => ",this.priceValue, this.stripeValue)
    }
}