import { Controller } from "stimulus"

export default class extends Controller {
    static targets = [ "email","password" ]

    validateEmail() {
        if (!this.emailTarget.validity.valid || this.emailTarget.value === ''){
            return false;
        }
        return true;
    }

    validatePassword() {
        if (this.passwordTarget.value === ''){
            return false;
        }
        return true;
    }


    submitForm(e){
        if (this.validateEmail() && this.validatePassword()){
            return;
        }
        e.preventDefault();

    }
}
