import {Injectable} from "@angular/core";
import {HttpClient, HttpHeaders} from "@angular/common/http";
import {Affordances} from "./affordance-structure";
import {Observable, Subscription} from "rxjs";

@Injectable({providedIn: "root"})
export class DataStorageService {
  private url = 'http://localhost:8080/upload';

  constructor(private http: HttpClient) {
  }

  fetchProtoData(file: File): Observable<Affordances> {
    const formData = new FormData();
    formData.append("uploadfile", file)

    const sendHeaders = new HttpHeaders({
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Credentials': 'true',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Allow-Methods': 'GET,PUT,POST,DELETE',
    });
    return this.http.post<Affordances>(this.url, formData);
  }

}
