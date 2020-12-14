import { Controller } from "stimulus"

export default class extends Controller {
    static values = { price: String }

    createCheckout(){
        console.log("price => ",this.priceValue)
    }
}