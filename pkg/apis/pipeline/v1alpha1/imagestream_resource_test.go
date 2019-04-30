package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_NewImagestreamResource(t *testing.T) {
	t.Run("valid imagestream resource", func(t *testing.T) {
		pr := &PipelineResource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "is-resource",
				Namespace: "test",
			},

			Spec: PipelineResourceSpec{
				Type: PipelineResourceTypeIS,
				Params: []Param{{
					Name:  "name",
					Value: "app:v1",
				}},
			},
		}

		want := &ImageStreamResource{
			Name: "app:v1",
			Type: PipelineResourceTypeIS,
			Ns:   "test",
		}

		got, err := NewImageStreamResource(pr)

		if err != nil {
			t.Errorf("Valid Imagestream resource. Error should be empty : %s\n", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Valid Imagestream resource.\nExpected %v\nGot %v\n", want, got)
		}
	})

	t.Run("invalid imagestream resource", func(t *testing.T) {
		pr := &PipelineResource{
			ObjectMeta: metav1.ObjectMeta{
				Name: "is-resource",
			},

			Spec: PipelineResourceSpec{
				Type: PipelineResourceTypeIS,
				Params: []Param{{
					Name:  "name",
					Value: "app:v1",
				}},
			},
		}

		_, err := NewImageStreamResource(pr)

		if err == nil {
			t.Errorf("Invalid Imagestream resource. Expecting error but its empty : %v\n", err)
		}

	})
}

func Test_ImagestreamResourceReplacement(t *testing.T) {
	pr := &PipelineResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "is-resource",
			Namespace: "test",
		},

		Spec: PipelineResourceSpec{
			Type: PipelineResourceTypeIS,
			Params: []Param{{
				Name:  "name",
				Value: "app:v1",
			}},
		},
	}

	want := map[string]string{
		"name": os4Registry + "/" + pr.Namespace + "/" + "app:v1",
		"type": string(PipelineResourceTypeIS),
	}

	isr, _ := NewImageStreamResource(pr)
	got := isr.Replacements()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Invalid replacement result.\nExpected %v\nGot %v\n", want, got)
	}
}
