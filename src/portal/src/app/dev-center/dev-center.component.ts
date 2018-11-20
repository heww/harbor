import { Component, OnInit } from "@angular/core";

@Component({
  selector: 'dev-center',
  templateUrl: './dev-center.component.html',
  styleUrls: ['./dev-center.component.scss']
})
export class DevCenterComponent {
    public sddc: any;
    public paramMap = new Map<string, string>();
    public VMC_SWAGGER_URL = "/path/to/swagger.json";
    public CSP_SWAGGER_URL = "/path/to/csp-swagger.json";

    constructor () {
        this.paramMap.set('org', 'some org id');
    }

    public selectSddc(sddc: any) {
        // swagger-list will only update with ngOnChanges, so must create new map
        let newMap = new Map(this.paramMap);
        newMap.set('sddc', sddc.value);
        this.paramMap = newMap;
    }
}
