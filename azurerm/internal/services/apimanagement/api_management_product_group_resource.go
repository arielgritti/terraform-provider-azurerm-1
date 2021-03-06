package apimanagement

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmApiManagementProductGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmApiManagementProductGroupCreate,
		Read:   resourceArmApiManagementProductGroupRead,
		Delete: resourceArmApiManagementProductGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"product_id": azure.SchemaApiManagementChildName(),

			"group_name": azure.SchemaApiManagementChildName(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"api_management_name": azure.SchemaApiManagementName(),
		},
	}
}

func resourceArmApiManagementProductGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.ProductGroupsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)
	serviceName := d.Get("api_management_name").(string)
	groupName := d.Get("group_name").(string)
	productId := d.Get("product_id").(string)

	exists, err := client.CheckEntityExists(ctx, resourceGroup, serviceName, productId, groupName)
	if err != nil {
		if !utils.ResponseWasNotFound(exists) {
			return fmt.Errorf("checking for present of existing Product %q / Group %q (API Management Service %q / Resource Group %q): %+v", productId, groupName, serviceName, resourceGroup, err)
		}
	}

	if !utils.ResponseWasNotFound(exists) {
		// TODO: can we pull this from somewhere?
		subscriptionId := meta.(*clients.Client).Account.SubscriptionId
		resourceId := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ApiManagement/service/%s/products/%s/groups/%s", subscriptionId, resourceGroup, serviceName, productId, groupName)
		return tf.ImportAsExistsError("azurerm_api_management_product_group", resourceId)
	}

	resp, err := client.CreateOrUpdate(ctx, resourceGroup, serviceName, productId, groupName)
	if err != nil {
		return fmt.Errorf("adding Product %q to Group %q (API Management Service %q / Resource Group %q): %+v", productId, groupName, serviceName, resourceGroup, err)
	}

	// there's no Read so this is best-effort
	d.SetId(*resp.ID)

	return resourceArmApiManagementProductGroupRead(d, meta)
}

func resourceArmApiManagementProductGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.ProductGroupsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	serviceName := id.Path["service"]
	groupName := id.Path["groups"]
	productId := id.Path["products"]

	resp, err := client.CheckEntityExists(ctx, resourceGroup, serviceName, productId, groupName)
	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			log.Printf("[DEBUG] Product %q was not found in Group %q (API Management Service %q / Resource Group %q) was not found - removing from state!", productId, groupName, serviceName, resourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Product %q / Group %q (API Management Service %q / Resource Group %q): %+v", productId, groupName, serviceName, resourceGroup, err)
	}

	d.Set("group_name", groupName)
	d.Set("product_id", productId)
	d.Set("resource_group_name", resourceGroup)
	d.Set("api_management_name", serviceName)

	return nil
}

func resourceArmApiManagementProductGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.ProductGroupsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	serviceName := id.Path["service"]
	groupName := id.Path["groups"]
	productId := id.Path["products"]

	if resp, err := client.Delete(ctx, resourceGroup, serviceName, productId, groupName); err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("removing Product %q from Group %q (API Management Service %q / Resource Group %q): %+v", productId, groupName, serviceName, resourceGroup, err)
		}
	}

	return nil
}
