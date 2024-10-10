package api

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetListI(t *testing.T) {
	testPods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod2",
			},
		},
	}
	cases := map[string]struct {
		call       func() (interface{}, error)
		wantError  error
		wantResult interface{}
	}{
		"nil wanting pods": {
			call: func() (interface{}, error) {
				r, err := GetItemsFromList[corev1.Pod](nil)
				return r, err
			},
			wantResult: ([]corev1.Pod)(nil),
			wantError:  cmpopts.AnyError,
		},

		"podlist to secret": {
			call: func() (interface{}, error) {
				r, err := GetItemsFromList[corev1.Secret](&corev1.PodList{Items: testPods})
				return r, err
			},
			wantResult: ([]corev1.Secret)(nil),
			wantError:  cmpopts.AnyError,
		},

		"podlist": {
			call: func() (interface{}, error) {
				r, err := GetItemsFromList[corev1.Pod](&corev1.PodList{Items: testPods})
				return r, err
			},
			wantResult: testPods,
			wantError:  nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gotResult, gotError := tc.call()

			if diff := cmp.Diff(tc.wantResult, gotResult); diff != "" {
				t.Errorf("unexpected result (want - / got +)\n%s", diff)
			}

			if diff := cmp.Diff(tc.wantError, gotError, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("unexpected error (want - / got +)\n%s", diff)
			}
		})
	}

}
