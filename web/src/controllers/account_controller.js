import { Controller } from "stimulus"

export default class extends Controller {
    static targets = [ "name","email", "password", "confirmPassword", "formError"]

    validateEmail() {
        if (!this.emailTarget.validity.valid || this.emailTarget.value === ''){
            return false;
        }
        return true;
    }


    validateName() {
        if (this.nameTarget.value === ''){
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

    submitForgetForm(e){
        if (this.validateEmail()){
            return;
        }
        e.preventDefault();
    }

    submitSignupForm(e){
        if (this.validateEmail() && this.validatePassword() && this.validateName()){
            return;
        }
        e.preventDefault();
    }

    submitLoginForm(e){
        if (this.validateEmail() && this.validatePassword()){
            return;
        }
        e.preventDefault();
    }

    submitResetForm(e){
        if (this.validatePassword() &&(this.passwordTarget.value === this.confirmPasswordTarget.value)){
            return;
        }

        this.showFormError("Password's don't match")
        e.preventDefault();
    }

    showFormError(message){
        this.formErrorTarget.classList.remove("is-hidden")
        this.formErrorTarget.innerHTML = message
    }

    hideFormError(){
        this.formErrorTarget.classList.add("is-hidden")
        e.preventDefault();
    }
}
