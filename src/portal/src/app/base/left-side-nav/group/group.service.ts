import {throwError as observableThrowError,  Observable} from "rxjs";
import {catchError, map} from 'rxjs/operators';
import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { UserGroup } from "./group";
import { CURRENT_BASE_HREF, HTTP_GET_OPTIONS, HTTP_JSON_OPTIONS } from "../../../shared/units/utils";

const userGroupEndpoint = CURRENT_BASE_HREF + "/usergroups";
const ldapGroupSearchEndpoint = CURRENT_BASE_HREF + "/ldap/groups/search?groupname=";

@Injectable({
  providedIn: 'root',
})
export class GroupService {
  constructor(private http: HttpClient) {}

  private handleErrorObservable(error: Response | any) {
    console.error(error.error || error);
    return observableThrowError(error.error || error);
  }

  getUserGroups(): Observable<UserGroup[]> {
    return this.http.get<UserGroup[]>(userGroupEndpoint, HTTP_GET_OPTIONS).pipe(
    map(response => {
      return response || [];
    }),
    catchError(error => {
      return this.handleErrorObservable(error);
    }), );
  }

  createGroup(group: UserGroup): Observable<any> {
    return this.http
      .post(userGroupEndpoint, group, HTTP_JSON_OPTIONS).pipe(
      map(response => {
        return response || [];
      }),
      catchError(this.handleErrorObservable), );
  }

  editGroup(group: UserGroup): Observable<any> {
    return this.http
    .put(`${userGroupEndpoint}/${group.id}`, group, HTTP_JSON_OPTIONS).pipe(
    map(response => {
      return response || [];
    }),
    catchError(this.handleErrorObservable), );
  }

  deleteGroup(group_id: number): Observable<any> {
    return this.http
    .delete(`${userGroupEndpoint}/${group_id}`).pipe(
    map(response => {
      return response || [];
    }),
    catchError(this.handleErrorObservable), );
  }

  searchGroup(group_name: string): Observable<UserGroup[]> {
    return this.http
    .get<UserGroup[]>(`${ldapGroupSearchEndpoint}${group_name}`, HTTP_GET_OPTIONS).pipe(
    map(response => {
      return response || [];
    }),
    catchError(this.handleErrorObservable), );
  }
}
