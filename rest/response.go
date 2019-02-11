package rest

import (
  "encoding/json"
  "log"
  "net/http"
  "os"
)

type PageInfo struct {
  // 1-based index
  PageIndex      int   `json:"pageIndex"`
  ItemsPerPage   int   `json:"itemsPerPage"`
  TotalItemCount int64 `json:"totalItemCount"`
  TotalPageCount int64 `json:"totalPageCount"`
}

type SearchParams struct {
  Scopes   []string  `json:"scopes"`
  Terms    []string  `json:"terms"`
  Sort     string    `json:"sort"`
  PageInfo *PageInfo `json:"pageInfo"`
}

func (sp *SearchParams) EnsureSingleScope() (RestError) {
  if len(sp.Scopes) > 1 {
    return BadRequestError("We currently only support a single scope.", nil)
  } else if (len(sp.Scopes) == 0) {
    return BadRequestError("No scope specified.", nil)
  } else {
    return nil
  }
}

func (sp *SearchParams) SetTotalPages(count int64) {
  var itemsPerPage int = sp.PageInfo.ItemsPerPage
  var pageIndex int = sp.PageInfo.PageIndex
  pageCount := count/int64(itemsPerPage)
  if count % int64(itemsPerPage) > 0 {
    pageCount += 1
  }
  sp.PageInfo = &PageInfo{PageIndex: int(pageIndex), ItemsPerPage: int(itemsPerPage), TotalItemCount: count, TotalPageCount: pageCount}
}

type standardResponse struct {
  Data         interface{}   `json:"data"`
  Message      string        `json:"message"`
  SearchParams *SearchParams `json:"searchParams,omitempty"`
}

func StandardResponse(w http.ResponseWriter, d interface{}, message string, searchParams *SearchParams) (error) {
  w.Header().Set("Content-Type", "application/json")

  resp := standardResponse{Data: d, Message: message, SearchParams: searchParams}

  var respBody []byte
  var err error
  if respBody, err = json.Marshal(resp); err != nil {
    restErr := ServerError("Could not format response.", err)
    HandleError(w, restErr)

    return restErr
  }
  w.Write(respBody)

  return nil
}

func HandleError(w http.ResponseWriter, err RestError) (RestError) {
  // Note that ultimately, we want to encode the error in JSON, but it was
  // proving problematic, so for now it's just text.
  if err.Code() == http.StatusInternalServerError {
    log.Printf("ERROR: %+v", err.Cause()) // Log server/untyped errors.
  } else if os.Getenv("NODE_ENV") != `production` {
    log.Printf("%+v", err)
    log.Printf("%+v", err.Cause())
  }

  // TODO: hide error and give reference number
  http.Error(w, err.Error(), err.Code())

  return err
}