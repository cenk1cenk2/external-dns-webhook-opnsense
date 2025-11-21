package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"

	"github.com/cenk1cenk2/external-dns-webhook-opnsense/api"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/interfaces"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/internal/services"
	"github.com/cenk1cenk2/external-dns-webhook-opnsense/test/fixtures"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("API", func() {
	Describe("API Service", func() {
		var a *api.Api

		BeforeEach(func() {
			c := fixtures.NewTestConfig()
			logger := fixtures.NewTestLogger()
			validator := services.NewValidator()

			a = api.NewApi(&api.ApiSvc{
				Log:       logger,
				Validator: validator,
			}, c.Api)
			Expect(a).ToNot(BeNil())
			Expect(a.Echo).ToNot(BeNil())
		})

		It("should be able to get to ready state when started", func() {
			go func() {
				defer GinkgoRecover()

				err := <-a.Start(":0")
				Expect(err).To(Equal(http.ErrServerClosed))
			}()

			<-a.IsReady()

			Expect(a.Shutdown()).ToNot(HaveOccurred())
		})

		It("should be able to create a new API", func() {
			go func() {
				defer GinkgoRecover()

				err := <-a.Start(":0")
				Expect(err).To(Equal(http.ErrServerClosed))
			}()

			Eventually(func() error {
				defer GinkgoRecover()

				addr := fmt.Sprintf("http://%s/", (<-a.GetListener()).Addr().String())
				GinkgoWriter.Printf("Trying address: %s\n", addr)
				_, err := http.Get(addr)

				if err != nil {
					GinkgoWriter.Printf("Got error: %s -> %w\n", reflect.TypeOf(err), err)
				}

				return err
			}, time.Second*3, time.Millisecond*100).
				ToNot(HaveOccurred())

			Expect(a.Shutdown()).ToNot(HaveOccurred())
		})
	})

	Describe("HTTP Error Handler", func() {
		var a *api.Api

		BeforeEach(func() {
			c := fixtures.NewTestConfig()
			logger := fixtures.NewTestLogger()
			validator := services.NewValidator()

			a = api.NewApi(&api.ApiSvc{
				Log:       logger,
				Validator: validator,
			}, c.Api)
			Expect(a).ToNot(BeNil())
		})

		It("should return a JSON response with the arbitatary error message", func() {
			c, res := fixtures.CreateEchoContext(a.Echo, httptest.NewRequest(http.MethodGet, "/", nil))

			a.Echo.HTTPErrorHandler(fmt.Errorf("test"), c)

			Expect(res.Code).To(Equal(http.StatusInternalServerError))

			GinkgoWriter.Printf(res.Body.String())

			body := fixtures.MustJsonUnmarshal(&interfaces.ApiError{}, res.Body.String())
			Expect(body.Status).To(Equal(http.StatusInternalServerError))
			Expect(body.Message).To(Equal("test"))
			Expect(body.Error()).To(Equal("test"))
		})

		It("should return a JSON response with the http error message", func() {
			c, res := fixtures.CreateEchoContext(a.Echo, httptest.NewRequest(http.MethodGet, "/", nil))

			a.Echo.HTTPErrorHandler(echo.NewHTTPError(http.StatusTeapot, fmt.Errorf("test")), c)

			Expect(res.Code).To(Equal(http.StatusTeapot))

			GinkgoWriter.Printf(res.Body.String())

			body := fixtures.MustJsonUnmarshal(&interfaces.ApiError{}, res.Body.String())
			Expect(body.Status).To(Equal(http.StatusTeapot))
			Expect(body.Message).To(Equal("test"))
			Expect(body.Error()).To(Equal("test"))
		})
	})

	Describe("Validator", func() {
		var a *api.Api

		BeforeEach(func() {
			c := fixtures.NewTestConfig()
			logger := fixtures.NewTestLogger()
			validator := services.NewValidator()

			a = api.NewApi(&api.ApiSvc{
				Log:       logger,
				Validator: validator,
			}, c.Api)
			Expect(a).ToNot(BeNil())
		})

		It("should be able to validate a valid struct in the echo context", func(ctx SpecContext) {
			type Data struct {
				Field string `validate:"required"`
			}

			d := &Data{
				Field: "test",
			}

			err := a.Echo.Validator.Validate(d)

			Expect(err).ToNot(HaveOccurred())
		})
	})
})
