**gomodest** is a demo app inspired from modest approaches to building webapps as enlisted in https://modestjs.works/. It can be used as a template to spin off simple Go webapps.

> The main idea is to use server rendered html with spots of client-side dynamism using SvelteJS & StimulusJS
 
The webapp is mostly plain old javascript, html, css but with spots of SvelteJS used for interactivity without page reloads. StimulusJS is used for sprinkling
interactivity in server rendered html & mounting Svelte components into divs.

A few things which were used:

1. Go, html/template, goview
2. SvelteJS
3. StimulusJS
4. Bulma CSS

Many more things in `go.mod` & `web/package.json`

To run, clone this repo and: 

```bash
$ cd web && yarn install && yarn watch
$ go run main.go
```

This snippet is the most interesting part of this demo: 

```html
{{define "content"}}
    <div class="columns is-mobile is-centered">
        <div class="column is-half-desktop">
            <div
                    data-target="svelte.component"
                    data-component-name="app"
                    data-component-props="{{.Data}}">
            </div>
        </div>
    </div>
    </div>
{{end}}
```

[source](https://github.com/adnaan/gomodest/blob/main/web/html/app.html)

In the above snippet, we use StimulusJS to mount a Svelte component by using the following code:

```js
import { Controller } from "stimulus";
     import components from "./components";
     
     export default class extends Controller {
         static targets = ["component"]
         connect() {
             if (this.componentTargets.length > 0){
                 this.componentTargets.forEach(el => {
                     const componentName = el.dataset.componentName;
                     const componentProps = el.dataset.componentProps ? JSON.parse(el.dataset.componentProps): {};
                     if (!(componentName in components)){
                         console.log(`svelte component: ${componentName}, not found!`)
                         return;
                     }
                     const app = new components[componentName]({
                         target: el,
                         props: componentProps
                     });
                 })
             }
         }
     }
```
[source](https://github.com/adnaan/gomodest/blob/main/web/src/controllers/svelte_controller.js)

This strategy allows us to mix server rendered HTML pages with client side dynamism.

Other possibly interesting aspects could be the layout of [web/html](https://github.com/adnaan/gomodest/tree/main/web/html) and the usage of the super nice [goview](https://github.com/foolin/goview) library to render html in these files: 

 1. [pkg/router.go](https://github.com/adnaan/gomodest/blob/main/pkg/router.go)
 2. [pkg/view.go](https://github.com/adnaan/gomodest/blob/main/pkg/view.go)
 3. [pkg/pages.go](https://github.com/adnaan/gomodest/blob/main/pkg/pages.go)
 
 That is all.
 
 ---------------
 
 Pages
 
![Home](screenshots/gomodest_4.png?raw=true "Home")

![Sign in](screenshots/gomodest_3.png?raw=true "Sign in")

![App](screenshots/gomodest_2.png?raw=true "App")

![Account](screenshots/gomodest_1.png?raw=true "Account")

 
 